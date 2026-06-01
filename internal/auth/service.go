package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidInput         = errors.New("invalid input")
	ErrTooManyLoginAttempts = errors.New("too many login attempts")
	ErrUnauthorized         = errors.New("unauthorized")
	ErrUserNotFound         = errors.New("user not found")
)

type RegisterInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResult struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type Service struct {
	repo         UserRepository
	jwtSecret    []byte
	tokenTTL     time.Duration
	loginLimiter LoginLimiter
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
}

type LoginLimiter interface {
	Check(ctx context.Context, email string) error
	RecordFailure(ctx context.Context, email string) error
	Reset(ctx context.Context, email string) error
}

type noopLoginLimiter struct{}

func (noopLoginLimiter) Check(context.Context, string) error {
	return nil
}

func (noopLoginLimiter) RecordFailure(context.Context, string) error {
	return nil
}

func (noopLoginLimiter) Reset(context.Context, string) error {
	return nil
}

type claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func NewService(repo UserRepository, jwtSecret string) *Service {
	return NewServiceWithLoginLimiter(repo, jwtSecret, noopLoginLimiter{})
}

func NewServiceWithLoginLimiter(repo UserRepository, jwtSecret string, limiter LoginLimiter) *Service {
	if limiter == nil {
		limiter = noopLoginLimiter{}
	}
	return &Service{
		repo:         repo,
		jwtSecret:    []byte(jwtSecret),
		tokenTTL:     24 * time.Hour,
		loginLimiter: limiter,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*UserResponse, error) {
	email := normalizeEmail(input.Email)
	name := strings.TrimSpace(input.Name)
	password := input.Password

	if email == "" || name == "" || len(password) < 8 || !strings.Contains(email, "@") {
		return nil, ErrInvalidInput
	}

	if _, err := s.repo.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyExists
	} else if !errors.Is(err, ErrUserNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &User{
		Email:        email,
		Name:         name,
		PasswordHash: string(hash),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	email := normalizeEmail(input.Email)
	if err := s.loginLimiter.Check(ctx, email); err != nil {
		return nil, err
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if errors.Is(err, ErrUserNotFound) {
		if recordErr := s.loginLimiter.RecordFailure(ctx, email); recordErr != nil {
			return nil, recordErr
		}
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		if recordErr := s.loginLimiter.RecordFailure(ctx, email); recordErr != nil {
			return nil, recordErr
		}
		return nil, ErrInvalidCredentials
	}

	if err := s.loginLimiter.Reset(ctx, email); err != nil {
		return nil, err
	}

	token, err := s.signToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

func (s *Service) CurrentUser(ctx context.Context, tokenString string) (*UserResponse, error) {
	userID, err := s.ParseToken(tokenString)
	if err != nil {
		return nil, ErrUnauthorized
	}

	return s.CurrentUserByID(ctx, userID)
}

func (s *Service) CurrentUserByID(ctx context.Context, userID uint) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if errors.Is(err, ErrUserNotFound) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

func (s *Service) signToken(userID uint) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})

	return token.SignedString(s.jwtSecret)
}

func (s *Service) ParseToken(tokenString string) (uint, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return 0, ErrUnauthorized
	}

	claims, ok := parsed.Claims.(*claims)
	if !ok || !parsed.Valid || claims.UserID == 0 {
		return 0, ErrUnauthorized
	}
	return claims.UserID, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
