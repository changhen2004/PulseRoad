package requirement

import (
	"context"
	"errors"
	"strings"

	"pulseroad/internal/product"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrInvalid   = errors.New("invalid input")
	ErrNotOwner  = errors.New("only the creator can delete this requirement")
)

type CreateRequirementInput struct {
	Title            string `json:"title"`
	Description      string `json:"description"`
	Priority         string `json:"priority"`
	SourceFeedbackID *uint  `json:"source_feedback_id"`
}

type UpdateRequirementInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
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

func (s *Service) Create(ctx context.Context, userID uint, productID uint, input CreateRequirementInput) (*RequirementResponse, error) {
	title := strings.TrimSpace(input.Title)
	description := strings.TrimSpace(input.Description)
	priority := strings.TrimSpace(input.Priority)
	if priority == "" {
		priority = PriorityP2
	}

	if userID == 0 || productID == 0 || title == "" || !validPriority(priority) {
		return nil, ErrInvalid
	}

	if _, err := s.productForUser(ctx, userID, productID); err != nil {
		return nil, err
	}

	req := &Requirement{
		ProductID:        productID,
		Title:            title,
		Description:      description,
		Status:           StatusPlanned,
		Priority:         priority,
		SourceFeedbackID: input.SourceFeedbackID,
		CreatedBy:        userID,
	}
	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	response := req.ToResponse()
	return &response, nil
}

func (s *Service) ListByProduct(ctx context.Context, userID uint, productID uint, status string, page int, pageSize int) (*RequirementPageResponse, error) {
	if userID == 0 || productID == 0 {
		return nil, ErrForbidden
	}
	status = strings.TrimSpace(status)
	if status != "" && !validStatus(status) {
		return nil, ErrInvalid
	}
	if _, err := s.productForUser(ctx, userID, productID); err != nil {
		return nil, err
	}
	items, total, err := s.repo.ListByProduct(ctx, productID, status, page, pageSize)
	if err != nil {
		return nil, err
	}
	response := make([]RequirementResponse, 0, len(items))
	for _, item := range items {
		response = append(response, item.ToResponse())
	}
	return &RequirementPageResponse{
		Items:    response,
		Page:     normalizePage(page),
		PageSize: normalizePageSize(pageSize),
		Total:    total,
	}, nil
}

func (s *Service) Get(ctx context.Context, userID uint, requirementID uint) (*RequirementResponse, error) {
	if userID == 0 || requirementID == 0 {
		return nil, ErrForbidden
	}
	req, err := s.repo.FindByID(ctx, requirementID)
	if err != nil {
		return nil, err
	}
	if _, err := s.productForUser(ctx, userID, req.ProductID); err != nil {
		return nil, err
	}
	response := req.ToResponse()
	return &response, nil
}

func (s *Service) Update(ctx context.Context, userID uint, requirementID uint, input UpdateRequirementInput) (*RequirementResponse, error) {
	if userID == 0 || requirementID == 0 {
		return nil, ErrForbidden
	}
	req, err := s.repo.FindByID(ctx, requirementID)
	if err != nil {
		return nil, err
	}
	if _, err := s.productForUser(ctx, userID, req.ProductID); err != nil {
		return nil, err
	}

	if title := strings.TrimSpace(input.Title); title != "" {
		req.Title = title
	}
	if description := strings.TrimSpace(input.Description); description != "" {
		req.Description = description
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		if !validStatus(status) {
			return nil, ErrInvalid
		}
		req.Status = status
	}
	if priority := strings.TrimSpace(input.Priority); priority != "" {
		if !validPriority(priority) {
			return nil, ErrInvalid
		}
		req.Priority = priority
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}
	response := req.ToResponse()
	return &response, nil
}

func (s *Service) Delete(ctx context.Context, userID uint, requirementID uint) error {
	if userID == 0 || requirementID == 0 {
		return ErrForbidden
	}
	req, err := s.repo.FindByID(ctx, requirementID)
	if err != nil {
		return err
	}
	if _, err := s.productForUser(ctx, userID, req.ProductID); err != nil {
		return err
	}
	if req.CreatedBy != userID {
		return ErrNotOwner
	}
	return s.repo.Delete(ctx, requirementID)
}

func (s *Service) productForUser(ctx context.Context, userID uint, productID uint) (*product.ProductResponse, error) {
	productResponse, err := s.productAccess.GetProduct(ctx, userID, productID)
	if errors.Is(err, product.ErrForbidden) {
		return nil, ErrForbidden
	}
	if errors.Is(err, product.ErrProductNotFound) {
		return nil, errors.New("product not found")
	}
	return productResponse, err
}
