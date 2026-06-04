package feedback

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pulseroad/internal/product"
)

var (
	ErrForbidden        = errors.New("forbidden")
	ErrInvalid          = errors.New("invalid input")
	ErrFeedbackNotFound = errors.New("feedback not found")
	ErrProductNotFound  = errors.New("product not found")
	ErrVoteExists       = errors.New("vote already exists")
	ErrVoteNotFound     = errors.New("vote not found")
)

type CreateFeedbackInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdateFeedbackStatusInput struct {
	Status string `json:"status"`
}

type ListFeedbackInput struct {
	Status   string
	Page     int
	PageSize int
}

type CreateCommentInput struct {
	Content string `json:"content"`
}

type ListFeedbackQuery struct {
	Status   string
	Page     int
	PageSize int
	UserID   uint
}

type FeedbackPage struct {
	Items    []Feedback
	Page     int
	PageSize int
	Total    int64
}

type RepositoryPort interface {
	Create(ctx context.Context, feedback *Feedback) error
	ListByProduct(ctx context.Context, productID uint) ([]Feedback, error)
	ListByProductPage(ctx context.Context, productID uint, query ListFeedbackQuery) (FeedbackPage, error)
	FindByID(ctx context.Context, id uint) (*Feedback, error)
	UpdateStatus(ctx context.Context, id uint, status string) (*Feedback, error)
	CreateComment(ctx context.Context, comment *FeedbackComment) error
	ListComments(ctx context.Context, feedbackID uint) ([]FeedbackComment, error)
	CreateVote(ctx context.Context, vote *FeedbackVote) error
	DeleteVote(ctx context.Context, feedbackID uint, userID uint) error
	CountVotes(ctx context.Context, feedbackID uint) (int64, error)
}

type ProductAccess interface {
	GetProduct(ctx context.Context, userID uint, productID uint) (*product.ProductResponse, error)
}

type FeedbackCreatedEvent struct {
	FeedbackID uint      `json:"feedback_id"`
	ProductID  uint      `json:"product_id"`
	TeamID     uint      `json:"team_id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	CreatedBy  uint      `json:"created_by"`
	OccurredAt time.Time `json:"occurred_at"`
}

type EventPublisher interface {
	PublishFeedbackCreated(ctx context.Context, event FeedbackCreatedEvent) error
}

type noopEventPublisher struct{}

func (noopEventPublisher) PublishFeedbackCreated(context.Context, FeedbackCreatedEvent) error {
	return nil
}

type Service struct {
	repo          RepositoryPort
	productAccess ProductAccess
	publisher     EventPublisher
}

func NewService(repo RepositoryPort, productAccess ProductAccess) *Service {
	return NewServiceWithPublisher(repo, productAccess, noopEventPublisher{})
}

func NewServiceWithPublisher(repo RepositoryPort, productAccess ProductAccess, publisher EventPublisher) *Service {
	if publisher == nil {
		publisher = noopEventPublisher{}
	}
	return &Service{repo: repo, productAccess: productAccess, publisher: publisher}
}

func (s *Service) CreateFeedback(ctx context.Context, userID uint, productID uint, input CreateFeedbackInput) (*FeedbackResponse, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if userID == 0 || productID == 0 || title == "" || content == "" {
		return nil, ErrInvalid
	}

	productResponse, err := s.productForUser(ctx, userID, productID)
	if err != nil {
		return nil, err
	}

	feedback := &Feedback{
		ProductID: productID,
		Title:     title,
		Content:   content,
		Status:    StatusOpen,
		CreatedBy: userID,
	}
	if err := s.repo.Create(ctx, feedback); err != nil {
		return nil, err
	}

	if err := s.publisher.PublishFeedbackCreated(ctx, FeedbackCreatedEvent{
		FeedbackID: feedback.ID,
		ProductID:  feedback.ProductID,
		TeamID:     productResponse.TeamID,
		Title:      feedback.Title,
		Status:     feedback.Status,
		CreatedBy:  feedback.CreatedBy,
		OccurredAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("publish feedback created event: %w", err)
	}

	response := feedback.ToResponse()
	return &response, nil
}

func (s *Service) ListFeedback(ctx context.Context, userID uint, productID uint) ([]FeedbackResponse, error) {
	if userID == 0 || productID == 0 {
		return nil, ErrForbidden
	}

	if err := s.requireProductAccess(ctx, userID, productID); err != nil {
		return nil, err
	}

	feedbackItems, err := s.repo.ListByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	response := make([]FeedbackResponse, 0, len(feedbackItems))
	for _, feedback := range feedbackItems {
		response = append(response, feedback.ToResponse())
	}
	return response, nil
}

func (s *Service) ListFeedbackPage(ctx context.Context, userID uint, productID uint, input ListFeedbackInput) (*FeedbackPageResponse, error) {
	if userID == 0 || productID == 0 {
		return nil, ErrForbidden
	}
	status := strings.TrimSpace(input.Status)
	if status != "" && !validStatus(status) {
		return nil, ErrInvalid
	}
	if err := s.requireProductAccess(ctx, userID, productID); err != nil {
		return nil, err
	}
	page := normalizePage(input.Page)
	pageSize := normalizePageSize(input.PageSize)
	feedbackPage, err := s.repo.ListByProductPage(ctx, productID, ListFeedbackQuery{
		Status:   status,
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
	})
	if err != nil {
		return nil, err
	}
	items := make([]FeedbackResponse, 0, len(feedbackPage.Items))
	for _, feedback := range feedbackPage.Items {
		items = append(items, feedback.ToResponse())
	}
	return &FeedbackPageResponse{
		Items:    items,
		Page:     feedbackPage.Page,
		PageSize: feedbackPage.PageSize,
		Total:    feedbackPage.Total,
	}, nil
}

func (s *Service) GetFeedback(ctx context.Context, userID uint, feedbackID uint) (*FeedbackResponse, error) {
	if userID == 0 || feedbackID == 0 {
		return nil, ErrForbidden
	}

	feedback, err := s.repo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, err
	}

	if err := s.requireProductAccess(ctx, userID, feedback.ProductID); err != nil {
		return nil, err
	}

	response := feedback.ToResponse()
	return &response, nil
}

func (s *Service) UpdateStatus(ctx context.Context, userID uint, feedbackID uint, input UpdateFeedbackStatusInput) (*FeedbackResponse, error) {
	if userID == 0 || feedbackID == 0 {
		return nil, ErrForbidden
	}

	status := strings.TrimSpace(input.Status)
	if !validStatus(status) {
		return nil, ErrInvalid
	}

	feedback, err := s.repo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, err
	}

	if err := s.requireProductAccess(ctx, userID, feedback.ProductID); err != nil {
		return nil, err
	}

	updated, err := s.repo.UpdateStatus(ctx, feedbackID, status)
	if err != nil {
		return nil, err
	}

	response := updated.ToResponse()
	return &response, nil
}

func (s *Service) CreateComment(ctx context.Context, userID uint, feedbackID uint, input CreateCommentInput) (*FeedbackCommentResponse, error) {
	content := strings.TrimSpace(input.Content)
	if userID == 0 || feedbackID == 0 || content == "" {
		return nil, ErrInvalid
	}
	feedback, err := s.feedbackForUser(ctx, userID, feedbackID)
	if err != nil {
		return nil, err
	}
	comment := &FeedbackComment{
		FeedbackID: feedback.ID,
		Content:    content,
		CreatedBy:  userID,
	}
	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}
	response := comment.ToResponse()
	return &response, nil
}

func (s *Service) ListComments(ctx context.Context, userID uint, feedbackID uint) ([]FeedbackCommentResponse, error) {
	feedback, err := s.feedbackForUser(ctx, userID, feedbackID)
	if err != nil {
		return nil, err
	}
	comments, err := s.repo.ListComments(ctx, feedback.ID)
	if err != nil {
		return nil, err
	}
	response := make([]FeedbackCommentResponse, 0, len(comments))
	for _, comment := range comments {
		response = append(response, comment.ToResponse())
	}
	return response, nil
}

func (s *Service) VoteFeedback(ctx context.Context, userID uint, feedbackID uint) (*FeedbackVoteResponse, error) {
	feedback, err := s.feedbackForUser(ctx, userID, feedbackID)
	if err != nil {
		return nil, err
	}
	err = s.repo.CreateVote(ctx, &FeedbackVote{FeedbackID: feedback.ID, UserID: userID})
	if err != nil && !errors.Is(err, ErrVoteExists) {
		return nil, err
	}
	return s.voteResponse(ctx, userID, *feedback, true)
}

func (s *Service) UnvoteFeedback(ctx context.Context, userID uint, feedbackID uint) (*FeedbackVoteResponse, error) {
	feedback, err := s.feedbackForUser(ctx, userID, feedbackID)
	if err != nil {
		return nil, err
	}
	err = s.repo.DeleteVote(ctx, feedback.ID, userID)
	if err != nil && !errors.Is(err, ErrVoteNotFound) {
		return nil, err
	}
	return s.voteResponse(ctx, userID, *feedback, false)
}

func (s *Service) feedbackForUser(ctx context.Context, userID uint, feedbackID uint) (*Feedback, error) {
	if userID == 0 || feedbackID == 0 {
		return nil, ErrForbidden
	}
	feedback, err := s.repo.FindByID(ctx, feedbackID)
	if err != nil {
		return nil, err
	}
	if err := s.requireProductAccess(ctx, userID, feedback.ProductID); err != nil {
		return nil, err
	}
	return feedback, nil
}

func (s *Service) voteResponse(ctx context.Context, userID uint, feedback Feedback, voted bool) (*FeedbackVoteResponse, error) {
	count, err := s.repo.CountVotes(ctx, feedback.ID)
	if err != nil {
		return nil, err
	}
	return &FeedbackVoteResponse{FeedbackID: feedback.ID, Voted: voted, VoteCount: count}, nil
}

func (s *Service) requireProductAccess(ctx context.Context, userID uint, productID uint) error {
	_, err := s.productForUser(ctx, userID, productID)
	return err
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

func validStatus(status string) bool {
	return status == StatusOpen || status == StatusResolved
}

func normalizePage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize < 1 {
		return 20
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}
