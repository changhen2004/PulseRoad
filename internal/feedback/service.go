package feedback

import (
	"context"
	"errors"
	"strings"

	"pulseroad/internal/product"
)

var (
	ErrForbidden        = errors.New("forbidden")
	ErrInvalid          = errors.New("invalid input")
	ErrFeedbackNotFound = errors.New("feedback not found")
	ErrProductNotFound  = errors.New("product not found")
)

type CreateFeedbackInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdateFeedbackStatusInput struct {
	Status string `json:"status"`
}

type RepositoryPort interface {
	Create(ctx context.Context, feedback *Feedback) error
	ListByProduct(ctx context.Context, productID uint) ([]Feedback, error)
	FindByID(ctx context.Context, id uint) (*Feedback, error)
	UpdateStatus(ctx context.Context, id uint, status string) (*Feedback, error)
}

type ProductAccess interface {
	GetProduct(ctx context.Context, userID uint, productID uint) (*product.ProductResponse, error)
}

type Service struct {
	repo          RepositoryPort
	productAccess ProductAccess
}

func NewService(repo RepositoryPort, productAccess ProductAccess) *Service {
	return &Service{repo: repo, productAccess: productAccess}
}

func (s *Service) CreateFeedback(ctx context.Context, userID uint, productID uint, input CreateFeedbackInput) (*FeedbackResponse, error) {
	title := strings.TrimSpace(input.Title)
	content := strings.TrimSpace(input.Content)
	if userID == 0 || productID == 0 || title == "" || content == "" {
		return nil, ErrInvalid
	}

	if err := s.requireProductAccess(ctx, userID, productID); err != nil {
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

func (s *Service) requireProductAccess(ctx context.Context, userID uint, productID uint) error {
	_, err := s.productAccess.GetProduct(ctx, userID, productID)
	if errors.Is(err, product.ErrForbidden) {
		return ErrForbidden
	}
	if errors.Is(err, product.ErrProductNotFound) {
		return ErrProductNotFound
	}
	return err
}

func validStatus(status string) bool {
	return status == StatusOpen || status == StatusResolved
}
