package flagflow

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"pulseroad/internal/product"
)

var (
	ErrCacheMiss         = errors.New("cache miss")
	ErrFlagAlreadyExists = errors.New("flag already exists")
	ErrFlagNotFound      = errors.New("flag not found")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalid           = errors.New("invalid input")
	ErrProductNotFound   = errors.New("product not found")
)

const (
	EvaluateReasonDisabled = "disabled"
	EvaluateReasonRollout  = "rollout"
	EvaluateReasonNotFound = "not_found"
)

type CreateFlagInput struct {
	Key               string `json:"key"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Environment       string `json:"environment"`
	RolloutPercentage int    `json:"rollout_percentage"`
}

type UpdateFlagInput struct {
	Key               string `json:"key"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	Environment       string `json:"environment"`
	RolloutPercentage int    `json:"rollout_percentage"`
}

type ToggleFlagInput struct {
	Enabled bool `json:"enabled"`
}

type EvaluateFlagInput struct {
	ProductID   uint   `json:"product_id"`
	Key         string `json:"key"`
	Environment string `json:"environment"`
	UserKey     string `json:"user_key"`
}

type EvaluateFlagResponse struct {
	Key               string `json:"key"`
	Environment       string `json:"environment"`
	Enabled           bool   `json:"enabled"`
	RolloutPercentage int    `json:"rollout_percentage"`
	Reason            string `json:"reason"`
}

type RepositoryPort interface {
	Create(ctx context.Context, flag *FeatureFlag) error
	ListByProduct(ctx context.Context, productID uint, environment string) ([]FeatureFlag, error)
	FindByID(ctx context.Context, id uint) (*FeatureFlag, error)
	FindByKey(ctx context.Context, productID uint, environment string, key string) (*FeatureFlag, error)
	Update(ctx context.Context, flag *FeatureFlag) error
}

type ProductAccess interface {
	GetProduct(ctx context.Context, userID uint, productID uint) (*product.ProductResponse, error)
}

type Cache interface {
	Get(ctx context.Context, productID uint, environment string, key string) (*FeatureFlag, error)
	Set(ctx context.Context, flag FeatureFlag) error
	Delete(ctx context.Context, productID uint, environment string, key string) error
}

type EventPublisher interface {
	PublishFlagChanged(ctx context.Context, event FlagChangedEvent) error
}

type FlagChangedEvent struct {
	FlagID            uint      `json:"flag_id"`
	ProductID         uint      `json:"product_id"`
	TeamID            uint      `json:"team_id"`
	Key               string    `json:"key"`
	Environment       string    `json:"environment"`
	Enabled           bool      `json:"enabled"`
	RolloutPercentage int       `json:"rollout_percentage"`
	Action            string    `json:"action"`
	ChangedBy         uint      `json:"changed_by"`
	OccurredAt        time.Time `json:"occurred_at"`
}

type noopCache struct{}

func (noopCache) Get(context.Context, uint, string, string) (*FeatureFlag, error) {
	return nil, ErrCacheMiss
}

func (noopCache) Set(context.Context, FeatureFlag) error {
	return nil
}

func (noopCache) Delete(context.Context, uint, string, string) error {
	return nil
}

type noopEventPublisher struct{}

func (noopEventPublisher) PublishFlagChanged(context.Context, FlagChangedEvent) error {
	return nil
}

type Service struct {
	repo          RepositoryPort
	productAccess ProductAccess
	cache         Cache
	publisher     EventPublisher
}

func NewService(repo RepositoryPort, productAccess ProductAccess, cache Cache, publisher EventPublisher) *Service {
	if cache == nil {
		cache = noopCache{}
	}
	if publisher == nil {
		publisher = noopEventPublisher{}
	}
	return &Service{repo: repo, productAccess: productAccess, cache: cache, publisher: publisher}
}

func (s *Service) CreateFlag(ctx context.Context, userID uint, productID uint, input CreateFlagInput) (*FeatureFlagResponse, error) {
	key := normalizeKey(input.Key)
	name := strings.TrimSpace(input.Name)
	environment := normalizeEnvironment(input.Environment)
	description := strings.TrimSpace(input.Description)
	if userID == 0 || productID == 0 || key == "" || name == "" || !validRollout(input.RolloutPercentage) {
		return nil, ErrInvalid
	}

	productResponse, err := s.productForUser(ctx, userID, productID)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.FindByKey(ctx, productID, environment, key); err == nil {
		return nil, ErrFlagAlreadyExists
	} else if !errors.Is(err, ErrFlagNotFound) {
		return nil, err
	}

	flag := &FeatureFlag{
		ProductID:         productID,
		Key:               key,
		Name:              name,
		Description:       description,
		Environment:       environment,
		Enabled:           false,
		RolloutPercentage: input.RolloutPercentage,
		CreatedBy:         userID,
	}
	if err := s.repo.Create(ctx, flag); err != nil {
		return nil, err
	}
	if err := s.publishChanged(ctx, *flag, productResponse.TeamID, userID, EventActionCreated); err != nil {
		return nil, err
	}

	response := flag.ToResponse()
	return &response, nil
}

func (s *Service) ListFlags(ctx context.Context, userID uint, productID uint, environment string) ([]FeatureFlagResponse, error) {
	if userID == 0 || productID == 0 {
		return nil, ErrForbidden
	}
	if _, err := s.productForUser(ctx, userID, productID); err != nil {
		return nil, err
	}
	flags, err := s.repo.ListByProduct(ctx, productID, normalizeOptionalEnvironment(environment))
	if err != nil {
		return nil, err
	}
	response := make([]FeatureFlagResponse, 0, len(flags))
	for _, flag := range flags {
		response = append(response, flag.ToResponse())
	}
	return response, nil
}

func (s *Service) GetFlag(ctx context.Context, userID uint, flagID uint) (*FeatureFlagResponse, error) {
	flag, _, err := s.flagForUser(ctx, userID, flagID)
	if err != nil {
		return nil, err
	}
	response := flag.ToResponse()
	return &response, nil
}

func (s *Service) UpdateFlag(ctx context.Context, userID uint, flagID uint, input UpdateFlagInput) (*FeatureFlagResponse, error) {
	flag, productResponse, err := s.flagForUser(ctx, userID, flagID)
	if err != nil {
		return nil, err
	}

	oldKey := flag.Key
	oldEnvironment := flag.Environment
	if input.Key != "" {
		flag.Key = normalizeKey(input.Key)
	}
	if input.Name != "" {
		flag.Name = strings.TrimSpace(input.Name)
	}
	flag.Description = strings.TrimSpace(input.Description)
	if input.Environment != "" {
		flag.Environment = normalizeEnvironment(input.Environment)
	}
	if !validRollout(input.RolloutPercentage) || flag.Key == "" || flag.Name == "" {
		return nil, ErrInvalid
	}
	flag.RolloutPercentage = input.RolloutPercentage

	if err := s.repo.Update(ctx, flag); err != nil {
		return nil, err
	}
	_ = s.cache.Delete(ctx, flag.ProductID, oldEnvironment, oldKey)
	_ = s.cache.Delete(ctx, flag.ProductID, flag.Environment, flag.Key)
	if err := s.publishChanged(ctx, *flag, productResponse.TeamID, userID, EventActionUpdated); err != nil {
		return nil, err
	}
	response := flag.ToResponse()
	return &response, nil
}

func (s *Service) ToggleFlag(ctx context.Context, userID uint, flagID uint, input ToggleFlagInput) (*FeatureFlagResponse, error) {
	flag, productResponse, err := s.flagForUser(ctx, userID, flagID)
	if err != nil {
		return nil, err
	}
	flag.Enabled = input.Enabled
	if err := s.repo.Update(ctx, flag); err != nil {
		return nil, err
	}
	_ = s.cache.Delete(ctx, flag.ProductID, flag.Environment, flag.Key)
	if err := s.publishChanged(ctx, *flag, productResponse.TeamID, userID, EventActionToggled); err != nil {
		return nil, err
	}
	response := flag.ToResponse()
	return &response, nil
}

func (s *Service) EvaluateFlag(ctx context.Context, userID uint, input EvaluateFlagInput) (*EvaluateFlagResponse, error) {
	key := normalizeKey(input.Key)
	environment := normalizeEnvironment(input.Environment)
	userKey := strings.TrimSpace(input.UserKey)
	if userID == 0 || input.ProductID == 0 || key == "" || userKey == "" {
		return nil, ErrInvalid
	}
	if _, err := s.productForUser(ctx, userID, input.ProductID); err != nil {
		return nil, err
	}

	flag, err := s.cache.Get(ctx, input.ProductID, environment, key)
	if errors.Is(err, ErrCacheMiss) {
		flag, err = s.repo.FindByKey(ctx, input.ProductID, environment, key)
		if errors.Is(err, ErrFlagNotFound) {
			return &EvaluateFlagResponse{Key: key, Environment: environment, Enabled: false, Reason: EvaluateReasonNotFound}, nil
		}
		if err != nil {
			return nil, err
		}
		_ = s.cache.Set(ctx, *flag)
	} else if err != nil {
		return nil, err
	}

	enabled, reason := evaluate(*flag, userKey)
	return &EvaluateFlagResponse{
		Key:               flag.Key,
		Environment:       flag.Environment,
		Enabled:           enabled,
		RolloutPercentage: flag.RolloutPercentage,
		Reason:            reason,
	}, nil
}

func (s *Service) flagForUser(ctx context.Context, userID uint, flagID uint) (*FeatureFlag, *product.ProductResponse, error) {
	if userID == 0 || flagID == 0 {
		return nil, nil, ErrForbidden
	}
	flag, err := s.repo.FindByID(ctx, flagID)
	if err != nil {
		return nil, nil, err
	}
	productResponse, err := s.productForUser(ctx, userID, flag.ProductID)
	if err != nil {
		return nil, nil, err
	}
	return flag, productResponse, nil
}

func (s *Service) productForUser(ctx context.Context, userID uint, productID uint) (*product.ProductResponse, error) {
	productResponse, err := s.productAccess.GetProduct(ctx, userID, productID)
	if errors.Is(err, product.ErrForbidden) {
		return nil, ErrForbidden
	}
	if errors.Is(err, product.ErrProductNotFound) {
		return nil, ErrProductNotFound
	}
	return productResponse, err
}

func (s *Service) publishChanged(ctx context.Context, flag FeatureFlag, teamID uint, userID uint, action string) error {
	if err := s.publisher.PublishFlagChanged(ctx, FlagChangedEvent{
		FlagID:            flag.ID,
		ProductID:         flag.ProductID,
		TeamID:            teamID,
		Key:               flag.Key,
		Environment:       flag.Environment,
		Enabled:           flag.Enabled,
		RolloutPercentage: flag.RolloutPercentage,
		Action:            action,
		ChangedBy:         userID,
		OccurredAt:        time.Now(),
	}); err != nil {
		return fmt.Errorf("publish flag changed event: %w", err)
	}
	return nil
}

func normalizeKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}

func normalizeEnvironment(environment string) string {
	normalized := strings.ToLower(strings.TrimSpace(environment))
	if normalized == "" {
		return DefaultEnvironment
	}
	return normalized
}

func normalizeOptionalEnvironment(environment string) string {
	return strings.ToLower(strings.TrimSpace(environment))
}

func validRollout(value int) bool {
	return value >= 0 && value <= 100
}

func evaluate(flag FeatureFlag, userKey string) (bool, string) {
	if !flag.Enabled {
		return false, EvaluateReasonDisabled
	}
	if flag.RolloutPercentage >= 100 {
		return true, EvaluateReasonRollout
	}
	if flag.RolloutPercentage <= 0 {
		return false, EvaluateReasonRollout
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(flag.Key + ":" + flag.Environment + ":" + userKey))
	return int(hash.Sum32()%100) < flag.RolloutPercentage, EvaluateReasonRollout
}
