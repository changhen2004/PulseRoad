package feedback

import (
	"context"
	"errors"
	"testing"
	"time"

	"pulseroad/internal/product"
)

type fakeFeedbackRepository struct {
	nextID   uint
	feedback map[uint]*Feedback
}

func newFakeFeedbackRepository() *fakeFeedbackRepository {
	return &fakeFeedbackRepository{
		nextID:   1,
		feedback: make(map[uint]*Feedback),
	}
}

func (r *fakeFeedbackRepository) Create(_ context.Context, feedback *Feedback) error {
	feedback.ID = r.nextID
	feedback.CreatedAt = time.Now()
	feedback.UpdatedAt = feedback.CreatedAt
	r.nextID++

	copy := *feedback
	r.feedback[feedback.ID] = &copy
	return nil
}

func (r *fakeFeedbackRepository) ListByProduct(_ context.Context, productID uint) ([]Feedback, error) {
	var items []Feedback
	for _, feedback := range r.feedback {
		if feedback.ProductID == productID {
			items = append(items, *feedback)
		}
	}
	return items, nil
}

func (r *fakeFeedbackRepository) FindByID(_ context.Context, id uint) (*Feedback, error) {
	feedback, ok := r.feedback[id]
	if !ok {
		return nil, ErrFeedbackNotFound
	}
	copy := *feedback
	return &copy, nil
}

func (r *fakeFeedbackRepository) UpdateStatus(_ context.Context, id uint, status string) (*Feedback, error) {
	feedback, ok := r.feedback[id]
	if !ok {
		return nil, ErrFeedbackNotFound
	}
	feedback.Status = status
	feedback.UpdatedAt = time.Now()
	copy := *feedback
	return &copy, nil
}

type fakeProductAccess struct {
	products map[uint]*product.ProductResponse
	members  map[uint]map[uint]bool
}

type fakeFeedbackEventPublisher struct {
	events []FeedbackCreatedEvent
	err    error
}

func (p *fakeFeedbackEventPublisher) PublishFeedbackCreated(_ context.Context, event FeedbackCreatedEvent) error {
	if p.err != nil {
		return p.err
	}
	p.events = append(p.events, event)
	return nil
}

func newFakeProductAccess() *fakeProductAccess {
	return &fakeProductAccess{
		products: make(map[uint]*product.ProductResponse),
		members:  make(map[uint]map[uint]bool),
	}
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

func TestCreateFeedbackRequiresProductMember(t *testing.T) {
	access := newFakeProductAccess()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	svc := NewService(newFakeFeedbackRepository(), access)

	created, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{
		Title:   "Missing export",
		Content: "CSV export would help.",
	})
	if err != nil {
		t.Fatalf("create feedback: %v", err)
	}

	if created.ID == 0 {
		t.Fatal("expected feedback id")
	}
	if created.ProductID != 10 {
		t.Fatalf("expected product id 10, got %d", created.ProductID)
	}
	if created.Status != StatusOpen || created.CreatedBy != 7 {
		t.Fatalf("unexpected feedback: %#v", created)
	}
	if created.Content != "CSV export would help." {
		t.Fatalf("expected content to be returned, got %q", created.Content)
	}
}

func TestCreateFeedbackPublishesCreatedEvent(t *testing.T) {
	access := newFakeProductAccess()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	publisher := &fakeFeedbackEventPublisher{}
	svc := NewServiceWithPublisher(newFakeFeedbackRepository(), access, publisher)

	created, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{
		Title:   "Missing export",
		Content: "CSV export would help.",
	})
	if err != nil {
		t.Fatalf("create feedback: %v", err)
	}

	if len(publisher.events) != 1 {
		t.Fatalf("expected one event, got %#v", publisher.events)
	}
	event := publisher.events[0]
	if event.FeedbackID != created.ID || event.ProductID != 10 || event.TeamID != 20 || event.CreatedBy != 7 {
		t.Fatalf("unexpected event: %#v", event)
	}
	if event.Title != "Missing export" || event.Status != StatusOpen {
		t.Fatalf("unexpected event payload: %#v", event)
	}
	if event.OccurredAt.IsZero() {
		t.Fatal("expected occurred_at to be set")
	}
}

func TestCreateFeedbackDoesNotPublishWhenCreateFails(t *testing.T) {
	access := newFakeProductAccess()
	access.addProduct(10, 20)
	publisher := &fakeFeedbackEventPublisher{}
	svc := NewServiceWithPublisher(newFakeFeedbackRepository(), access, publisher)

	_, err := svc.CreateFeedback(context.Background(), 8, 10, CreateFeedbackInput{
		Title:   "Missing export",
		Content: "CSV export would help.",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if len(publisher.events) != 0 {
		t.Fatalf("expected no events, got %#v", publisher.events)
	}
}

func TestCreateFeedbackRejectsNonMember(t *testing.T) {
	access := newFakeProductAccess()
	access.addProduct(10, 20)
	svc := NewService(newFakeFeedbackRepository(), access)

	_, err := svc.CreateFeedback(context.Background(), 8, 10, CreateFeedbackInput{
		Title:   "Missing export",
		Content: "CSV export would help.",
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCreateFeedbackRejectsMissingProduct(t *testing.T) {
	svc := NewService(newFakeFeedbackRepository(), newFakeProductAccess())

	_, err := svc.CreateFeedback(context.Background(), 7, 404, CreateFeedbackInput{
		Title:   "Missing export",
		Content: "CSV export would help.",
	})
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
}

func TestListAndGetFeedbackRequiresProductMember(t *testing.T) {
	access := newFakeProductAccess()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	repo := newFakeFeedbackRepository()
	svc := NewService(repo, access)

	created, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{
		Title:   "Missing export",
		Content: "CSV export would help.",
	})
	if err != nil {
		t.Fatalf("create feedback: %v", err)
	}
	if _, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{
		Title:   "Dark mode",
		Content: "Dark mode would reduce eye strain.",
	}); err != nil {
		t.Fatalf("create second feedback: %v", err)
	}

	items, err := svc.ListFeedback(context.Background(), 7, 10)
	if err != nil {
		t.Fatalf("list feedback: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected two feedback items, got %#v", items)
	}

	got, err := svc.GetFeedback(context.Background(), 7, created.ID)
	if err != nil {
		t.Fatalf("get feedback: %v", err)
	}
	if got.ID != created.ID || got.ProductID != 10 {
		t.Fatalf("unexpected feedback detail: %#v", got)
	}
	if got.Content != "CSV export would help." {
		t.Fatalf("expected content to round-trip, got %q", got.Content)
	}
}

func TestGetFeedbackRejectsNonMember(t *testing.T) {
	access := newFakeProductAccess()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	repo := newFakeFeedbackRepository()
	svc := NewService(repo, access)

	created, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{
		Title:   "Missing export",
		Content: "CSV export would help.",
	})
	if err != nil {
		t.Fatalf("create feedback: %v", err)
	}

	_, err = svc.GetFeedback(context.Background(), 8, created.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateStatus(t *testing.T) {
	access := newFakeProductAccess()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	repo := newFakeFeedbackRepository()
	svc := NewService(repo, access)

	created, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{
		Title:   "Missing export",
		Content: "CSV export would help.",
	})
	if err != nil {
		t.Fatalf("create feedback: %v", err)
	}

	updated, err := svc.UpdateStatus(context.Background(), 7, created.ID, UpdateFeedbackStatusInput{Status: StatusResolved})
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	if updated.Status != StatusResolved {
		t.Fatalf("expected status resolved, got %q", updated.Status)
	}
	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Fatalf("expected updated_at after original timestamp, got original=%s updated=%s", created.UpdatedAt, updated.UpdatedAt)
	}

	got, err := svc.GetFeedback(context.Background(), 7, created.ID)
	if err != nil {
		t.Fatalf("get feedback after update: %v", err)
	}
	if got.Status != StatusResolved {
		t.Fatalf("expected persisted status resolved, got %q", got.Status)
	}

	updatedAgain, err := svc.UpdateStatus(context.Background(), 7, created.ID, UpdateFeedbackStatusInput{Status: StatusResolved})
	if err != nil {
		t.Fatalf("repeat update status: %v", err)
	}
	if updatedAgain.Status != StatusResolved {
		t.Fatalf("expected repeated status resolved, got %q", updatedAgain.Status)
	}

	_, err = svc.UpdateStatus(context.Background(), 7, created.ID, UpdateFeedbackStatusInput{Status: "closed"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
