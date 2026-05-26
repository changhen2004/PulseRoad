package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const testJWTSecret = "test-secret-that-is-long-enough-for-hs256"

type fakeUserRepository struct {
	nextID uint
	users  map[uint]*User
	emails map[string]uint
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		nextID: 1,
		users:  make(map[uint]*User),
		emails: make(map[string]uint),
	}
}

func (r *fakeUserRepository) Create(_ context.Context, user *User) error {
	if _, ok := r.emails[user.Email]; ok {
		return ErrEmailAlreadyExists
	}
	now := time.Now()
	user.ID = r.nextID
	user.CreatedAt = now
	user.UpdatedAt = now
	r.nextID++

	copy := *user
	r.users[user.ID] = &copy
	r.emails[user.Email] = user.ID
	return nil
}

func (r *fakeUserRepository) FindByEmail(_ context.Context, email string) (*User, error) {
	id, ok := r.emails[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	copy := *r.users[id]
	return &copy, nil
}

func (r *fakeUserRepository) FindByID(_ context.Context, id uint) (*User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	copy := *user
	return &copy, nil
}

func newTestService(t *testing.T) (*Service, *fakeUserRepository) {
	t.Helper()

	repo := newFakeUserRepository()
	return NewService(repo, testJWTSecret), repo
}

func TestRegisterStoresBcryptHash(t *testing.T) {
	svc, repo := newTestService(t)

	user, err := svc.Register(context.Background(), RegisterInput{
		Email:    "ada@example.com",
		Password: "correct horse battery staple",
		Name:     "Ada Lovelace",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	stored, err := repo.FindByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("find stored user: %v", err)
	}

	if stored.PasswordHash == "correct horse battery staple" {
		t.Fatal("password was stored in plaintext")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("correct horse battery staple")); err != nil {
		t.Fatalf("stored password hash is not a bcrypt hash for the password: %v", err)
	}
	if user.Email != "ada@example.com" || user.Name != "Ada Lovelace" {
		t.Fatalf("unexpected user response: %#v", user)
	}
}

func TestRegisterRejectsDuplicateEmail(t *testing.T) {
	svc, _ := newTestService(t)
	input := RegisterInput{Email: "ada@example.com", Password: "password123", Name: "Ada"}

	if _, err := svc.Register(context.Background(), input); err != nil {
		t.Fatalf("first register: %v", err)
	}
	_, err := svc.Register(context.Background(), input)
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Fatalf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestLoginReturnsTokenForValidPassword(t *testing.T) {
	svc, _ := newTestService(t)
	registered, err := svc.Register(context.Background(), RegisterInput{
		Email:    "ada@example.com",
		Password: "password123",
		Name:     "Ada",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	result, err := svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected token")
	}
	if result.User.ID != registered.ID {
		t.Fatalf("expected logged in user %d, got %d", registered.ID, result.User.ID)
	}

	current, err := svc.CurrentUser(context.Background(), result.Token)
	if err != nil {
		t.Fatalf("current user from token: %v", err)
	}
	if current.ID != registered.ID {
		t.Fatalf("expected current user %d, got %d", registered.ID, current.ID)
	}

	userID, err := svc.ParseToken(result.Token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if userID != registered.ID {
		t.Fatalf("expected parsed user id %d, got %d", registered.ID, userID)
	}
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	svc, _ := newTestService(t)
	if _, err := svc.Register(context.Background(), RegisterInput{
		Email:    "ada@example.com",
		Password: "password123",
		Name:     "Ada",
	}); err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err := svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	svc, _ := newTestService(t)
	svc.tokenTTL = -time.Hour

	registered, err := svc.Register(context.Background(), RegisterInput{
		Email:    "ada@example.com",
		Password: "password123",
		Name:     "Ada",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	result, err := svc.Login(context.Background(), LoginInput{
		Email:    registered.Email,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	_, err = svc.ParseToken(result.Token)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}
