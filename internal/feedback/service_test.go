package feedback

import (
	"context"
	"errors"
	"testing"
	"time"

	"pulseroad/internal/product"
)

type fakeFeedbackRepository struct {
	nextID        uint
	nextCommentID uint
	feedback      map[uint]*Feedback
	comments      map[uint][]FeedbackComment
	votes         map[uint]map[uint]*FeedbackVote
}

func newFakeFeedbackRepository() *fakeFeedbackRepository {
	return &fakeFeedbackRepository{
		nextID:        1,
		nextCommentID: 1,
		feedback:      make(map[uint]*Feedback),
		comments:      make(map[uint][]FeedbackComment),
		votes:         make(map[uint]map[uint]*FeedbackVote),
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

func (r *fakeFeedbackRepository) ListByProductPage(_ context.Context, productID uint, query ListFeedbackQuery) (FeedbackPage, error) {
	var filtered []Feedback
	for _, feedback := range r.feedback {
		if feedback.ProductID != productID {
			continue
		}
		if query.Status != "" && feedback.Status != query.Status {
			continue
		}
		item := *feedback
		item.CommentCount = int64(len(r.comments[feedback.ID]))
		item.VoteCount = int64(len(r.votes[feedback.ID]))
		if r.votes[feedback.ID] != nil && r.votes[feedback.ID][query.UserID] != nil {
			item.Voted = true
		}
		filtered = append(filtered, item)
	}
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	return FeedbackPage{Items: filtered[start:end], Total: int64(len(filtered)), Page: page, PageSize: pageSize}, nil
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

func (r *fakeFeedbackRepository) CreateComment(_ context.Context, comment *FeedbackComment) error {
	comment.ID = r.nextCommentID
	comment.CreatedAt = time.Now()
	r.nextCommentID++
	copyComment := *comment
	r.comments[comment.FeedbackID] = append(r.comments[comment.FeedbackID], copyComment)
	return nil
}

func (r *fakeFeedbackRepository) ListComments(_ context.Context, feedbackID uint) ([]FeedbackComment, error) {
	return append([]FeedbackComment(nil), r.comments[feedbackID]...), nil
}

func (r *fakeFeedbackRepository) CreateVote(_ context.Context, vote *FeedbackVote) error {
	if r.votes[vote.FeedbackID] == nil {
		r.votes[vote.FeedbackID] = make(map[uint]*FeedbackVote)
	}
	if _, ok := r.votes[vote.FeedbackID][vote.UserID]; ok {
		return ErrVoteExists
	}
	vote.ID = uint(len(r.votes[vote.FeedbackID]) + 1)
	vote.CreatedAt = time.Now()
	copyVote := *vote
	r.votes[vote.FeedbackID][vote.UserID] = &copyVote
	return nil
}

func (r *fakeFeedbackRepository) DeleteVote(_ context.Context, feedbackID uint, userID uint) error {
	if r.votes[feedbackID] == nil || r.votes[feedbackID][userID] == nil {
		return ErrVoteNotFound
	}
	delete(r.votes[feedbackID], userID)
	return nil
}

func (r *fakeFeedbackRepository) CountVotes(_ context.Context, feedbackID uint) (int64, error) {
	return int64(len(r.votes[feedbackID])), nil
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

func TestListFeedbackSupportsStatusFilterPaginationAndVoteMetadata(t *testing.T) {
	access := newFakeProductAccess()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	repo := newFakeFeedbackRepository()
	svc := NewService(repo, access)

	first, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{Title: "A", Content: "A content"})
	if err != nil {
		t.Fatalf("create first feedback: %v", err)
	}
	if _, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{Title: "B", Content: "B content"}); err != nil {
		t.Fatalf("create second feedback: %v", err)
	}
	if _, err := svc.UpdateStatus(context.Background(), 7, first.ID, UpdateFeedbackStatusInput{Status: StatusResolved}); err != nil {
		t.Fatalf("resolve first feedback: %v", err)
	}
	if _, err := svc.VoteFeedback(context.Background(), 7, first.ID); err != nil {
		t.Fatalf("vote feedback: %v", err)
	}

	page, err := svc.ListFeedbackPage(context.Background(), 7, 10, ListFeedbackInput{Status: StatusResolved, Page: 1, PageSize: 1})
	if err != nil {
		t.Fatalf("list feedback page: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != first.ID {
		t.Fatalf("unexpected page: %#v", page)
	}
	if page.Items[0].VoteCount != 1 || !page.Items[0].Voted {
		t.Fatalf("expected vote metadata, got %#v", page.Items[0])
	}
}

func TestCreateAndListCommentsRequireProductMember(t *testing.T) {
	access := newFakeProductAccess()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	repo := newFakeFeedbackRepository()
	svc := NewService(repo, access)
	created, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{Title: "A", Content: "A content"})
	if err != nil {
		t.Fatalf("create feedback: %v", err)
	}

	comment, err := svc.CreateComment(context.Background(), 7, created.ID, CreateCommentInput{Content: "I need this too"})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if comment.ID == 0 || comment.Content != "I need this too" || comment.CreatedBy != 7 {
		t.Fatalf("unexpected comment: %#v", comment)
	}

	comments, err := svc.ListComments(context.Background(), 7, created.ID)
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if len(comments) != 1 || comments[0].ID != comment.ID {
		t.Fatalf("unexpected comments: %#v", comments)
	}
}

func TestVoteFeedbackIsIdempotentAndCanBeCanceled(t *testing.T) {
	access := newFakeProductAccess()
	access.addProduct(10, 20)
	access.addMember(10, 7)
	repo := newFakeFeedbackRepository()
	svc := NewService(repo, access)
	created, err := svc.CreateFeedback(context.Background(), 7, 10, CreateFeedbackInput{Title: "A", Content: "A content"})
	if err != nil {
		t.Fatalf("create feedback: %v", err)
	}

	result, err := svc.VoteFeedback(context.Background(), 7, created.ID)
	if err != nil {
		t.Fatalf("vote feedback: %v", err)
	}
	if !result.Voted || result.VoteCount != 1 {
		t.Fatalf("expected voted result, got %#v", result)
	}
	result, err = svc.VoteFeedback(context.Background(), 7, created.ID)
	if err != nil {
		t.Fatalf("repeat vote feedback: %v", err)
	}
	if !result.Voted || result.VoteCount != 1 {
		t.Fatalf("expected idempotent voted result, got %#v", result)
	}
	result, err = svc.UnvoteFeedback(context.Background(), 7, created.ID)
	if err != nil {
		t.Fatalf("unvote feedback: %v", err)
	}
	if result.Voted || result.VoteCount != 0 {
		t.Fatalf("expected canceled vote result, got %#v", result)
	}
}
