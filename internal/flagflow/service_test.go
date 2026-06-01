package flagflow

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"pulseroad/internal/product"
)

type fakeFlagRepository struct {
	nextID uint
	flags  map[uint]*FeatureFlag
}

func newFakeFlagRepository() *fakeFlagRepository {
	return &fakeFlagRepository{nextID: 1, flags: make(map[uint]*FeatureFlag)}
}

func (r *fakeFlagRepository) Create(_ context.Context, flag *FeatureFlag) error {
	for _, existing := range r.flags {
		if existing.ProductID == flag.ProductID && existing.Environment == flag.Environment && existing.Key == flag.Key {
			return ErrFlagAlreadyExists
		}
	}
	now := time.Now()
	flag.ID = r.nextID
	flag.CreatedAt = now
	flag.UpdatedAt = now
	r.nextID++
	copy := *flag
	r.flags[flag.ID] = &copy
	return nil
}

func (r *fakeFlagRepository) ListByProduct(_ context.Context, productID uint, environment string) ([]FeatureFlag, error) {
	var flags []FeatureFlag
	for _, flag := range r.flags {
		if flag.ProductID == productID && (environment == "" || flag.Environment == environment) {
			flags = append(flags, *flag)
		}
	}
	return flags, nil
}

func (r *fakeFlagRepository) FindByID(_ context.Context, id uint) (*FeatureFlag, error) {
	flag, ok := r.flags[id]
	if !ok {
		return nil, ErrFlagNotFound
	}
	copy := *flag
	return &copy, nil
}

func (r *fakeFlagRepository) FindByKey(_ context.Context, productID uint, environment string, key string) (*FeatureFlag, error) {
	for _, flag := range r.flags {
		if flag.ProductID == productID && flag.Environment == environment && flag.Key == key {
			copy := *flag
			return &copy, nil
		}
	}
	return nil, ErrFlagNotFound
}

func (r *fakeFlagRepository) Update(_ context.Context, flag *FeatureFlag) error {
	existing, ok := r.flags[flag.ID]
	if !ok {
		return ErrFlagNotFound
	}
	for _, other := range r.flags {
		if other.ID != flag.ID && other.ProductID == flag.ProductID && other.Environment == flag.Environment && other.Key == flag.Key {
			return ErrFlagAlreadyExists
		}
	}
	flag.CreatedAt = existing.CreatedAt
	flag.UpdatedAt = time.Now()
	copy := *flag
	r.flags[flag.ID] = &copy
	return nil
}

type fakeProductAccess struct {
	products map[uint]*product.ProductResponse
	members  map[uint]map[uint]bool
}

func newFakeProductAccess() *fakeProductAccess {
	return &fakeProductAccess{products: make(map[uint]*product.ProductResponse), members: make(map[uint]map[uint]bool)}
}

func (a *fakeProductAccess) addProduct(productID uint, teamID uint) {
	a.products[productID] = &product.ProductResponse{ID: productID, TeamID: teamID}
}

func (a *fakeProductAccess) addMember(productID uint, userID uint) {
	if a.members[productID] == nil {
		a.members[productID] = make(map[uint]bool)
	}
	a.members[productID][userID] = true
}

func (a *fakeProductAccess) GetProduct(_ context.Context, userID uint, productID uint) (*product.ProductResponse, error) {
	productResponse, ok := a.products[productID]
	if !ok {
		return nil, product.ErrProductNotFound
	}
	if !a.members[productID][userID] {
		return nil, product.ErrForbidden
	}
	copy := *productResponse
	return &copy, nil
}

type fakeCache struct {
	flags   map[string]*FeatureFlag
	deleted []string
}

func newFakeCache() *fakeCache {
	return &fakeCache{flags: make(map[string]*FeatureFlag)}
}

func (c *fakeCache) Get(_ context.Context, productID uint, environment string, key string) (*FeatureFlag, error) {
	flag, ok := c.flags[cacheKey(productID, environment, key)]
	if !ok {
		return nil, ErrCacheMiss
	}
	copy := *flag
	return &copy, nil
}

func (c *fakeCache) Set(_ context.Context, flag FeatureFlag) error {
	copy := flag
	c.flags[cacheKey(flag.ProductID, flag.Environment, flag.Key)] = &copy
	return nil
}

func (c *fakeCache) Delete(_ context.Context, productID uint, environment string, key string) error {
	c.deleted = append(c.deleted, cacheKey(productID, environment, key))
	delete(c.flags, cacheKey(productID, environment, key))
	return nil
}

type fakeEventPublisher struct {
	events []FlagChangedEvent
}

func (p *fakeEventPublisher) PublishFlagChanged(_ context.Context, event FlagChangedEvent) error {
	p.events = append(p.events, event)
	return nil
}

func testService() (*Service, *fakeFlagRepository, *fakeProductAccess, *fakeCache, *fakeEventPublisher) {
	repo := newFakeFlagRepository()
	access := newFakeProductAccess()
	cache := newFakeCache()
	publisher := &fakeEventPublisher{}
	return NewService(repo, access, cache, publisher), repo, access, cache, publisher
}

func TestCreateFlagRequiresProductMember(t *testing.T) {
	svc, _, access, _, publisher := testService()
	access.addProduct(10, 20)
	access.addMember(10, 7)

	created, err := svc.CreateFlag(context.Background(), 7, 10, CreateFlagInput{
		Key:               "new_dashboard",
		Name:              "New Dashboard",
		Description:       "Roll out the new dashboard",
		Environment:       "production",
		RolloutPercentage: 30,
	})
	if err != nil {
		t.Fatalf("create flag: %v", err)
	}

	if created.ID == 0 || created.ProductID != 10 || created.CreatedBy != 7 {
		t.Fatalf("unexpected flag: %#v", created)
	}
	if created.Enabled {
		t.Fatal("new flag should default to disabled")
	}
	if created.Environment != "production" || created.RolloutPercentage != 30 {
		t.Fatalf("unexpected rollout fields: %#v", created)
	}
	if len(publisher.events) != 1 || publisher.events[0].Action != EventActionCreated {
		t.Fatalf("expected created event, got %#v", publisher.events)
	}
}

func TestCreateFlagRejectsNonMember(t *testing.T) {
	svc, _, access, _, _ := testService()
	access.addProduct(10, 20)

	_, err := svc.CreateFlag(context.Background(), 8, 10, CreateFlagInput{
		Key:         "new_dashboard",
		Name:        "New Dashboard",
		Environment: "production",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCreateFlagRejectsDuplicateKeyInEnvironment(t *testing.T) {
	svc, _, access, _, _ := testService()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	input := CreateFlagInput{Key: "new_dashboard", Name: "New Dashboard", Environment: "production"}

	if _, err := svc.CreateFlag(context.Background(), 7, 10, input); err != nil {
		t.Fatalf("create first flag: %v", err)
	}
	_, err := svc.CreateFlag(context.Background(), 7, 10, input)
	if !errors.Is(err, ErrFlagAlreadyExists) {
		t.Fatalf("expected ErrFlagAlreadyExists, got %v", err)
	}
}

func TestUpdateAndToggleFlagInvalidateCacheAndPublishEvents(t *testing.T) {
	svc, _, access, cache, publisher := testService()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	created, err := svc.CreateFlag(context.Background(), 7, 10, CreateFlagInput{
		Key:         "new_dashboard",
		Name:        "New Dashboard",
		Environment: "production",
	})
	if err != nil {
		t.Fatalf("create flag: %v", err)
	}
	if err := cache.Set(context.Background(), FeatureFlag{ProductID: 10, Key: "new_dashboard", Environment: "production"}); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	updated, err := svc.UpdateFlag(context.Background(), 7, created.ID, UpdateFlagInput{
		Name:              "New Dashboard UI",
		Description:       "Updated",
		RolloutPercentage: 75,
	})
	if err != nil {
		t.Fatalf("update flag: %v", err)
	}
	if updated.Name != "New Dashboard UI" || updated.RolloutPercentage != 75 {
		t.Fatalf("unexpected update: %#v", updated)
	}

	toggled, err := svc.ToggleFlag(context.Background(), 7, created.ID, ToggleFlagInput{Enabled: true})
	if err != nil {
		t.Fatalf("toggle flag: %v", err)
	}
	if !toggled.Enabled {
		t.Fatal("expected flag enabled")
	}
	if len(cache.deleted) < 2 {
		t.Fatalf("expected cache invalidation on update and toggle, got %#v", cache.deleted)
	}
	if publisher.events[len(publisher.events)-1].Action != EventActionToggled {
		t.Fatalf("expected toggled event, got %#v", publisher.events)
	}
}

func TestEvaluateFlagUsesCacheAndRolloutPercentage(t *testing.T) {
	svc, _, access, cache, _ := testService()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	if err := cache.Set(context.Background(), FeatureFlag{
		ID:                1,
		ProductID:         10,
		Key:               "new_dashboard",
		Name:              "New Dashboard",
		Environment:       "production",
		Enabled:           true,
		RolloutPercentage: 100,
	}); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	result, err := svc.EvaluateFlag(context.Background(), 7, EvaluateFlagInput{
		ProductID:   10,
		Key:         "new_dashboard",
		Environment: "production",
		UserKey:     "user-10001",
	})
	if err != nil {
		t.Fatalf("evaluate flag: %v", err)
	}
	if !result.Enabled || result.Reason != EvaluateReasonRollout {
		t.Fatalf("expected enabled rollout result, got %#v", result)
	}
}

func TestEvaluateFlagReturnsFalseWhenDisabled(t *testing.T) {
	svc, _, access, cache, _ := testService()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	if err := cache.Set(context.Background(), FeatureFlag{
		ID:          1,
		ProductID:   10,
		Key:         "new_dashboard",
		Environment: "production",
		Enabled:     false,
	}); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	result, err := svc.EvaluateFlag(context.Background(), 7, EvaluateFlagInput{
		ProductID:   10,
		Key:         "new_dashboard",
		Environment: "production",
		UserKey:     "user-10001",
	})
	if err != nil {
		t.Fatalf("evaluate flag: %v", err)
	}
	if result.Enabled || result.Reason != EvaluateReasonDisabled {
		t.Fatalf("expected disabled result, got %#v", result)
	}
}

func cacheKey(productID uint, environment string, key string) string {
	return fmt.Sprintf("%d:%s:%s", productID, environment, key)
}
