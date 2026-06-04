package product

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrForbidden       = errors.New("forbidden")
	ErrInvalid         = errors.New("invalid input")
	ErrProductNotFound = errors.New("product not found")
)

type CreateProductInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RepositoryPort interface {
	Create(ctx context.Context, product *Product) error
	ListByTeam(ctx context.Context, teamID uint) ([]Product, error)
	FindByID(ctx context.Context, id uint) (*Product, error)
	SummaryStats(ctx context.Context, productID uint) (*ProductSummaryStats, error)
}

type TeamMembership interface {
	IsMember(ctx context.Context, userID uint, teamID uint) (bool, error)
}

type Service struct {
	repo       RepositoryPort
	membership TeamMembership
}

func NewService(repo RepositoryPort, membership TeamMembership) *Service {
	return &Service{repo: repo, membership: membership}
}

func (s *Service) CreateProduct(ctx context.Context, userID uint, teamID uint, input CreateProductInput) (*ProductResponse, error) {
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)
	if userID == 0 || teamID == 0 || name == "" {
		return nil, ErrInvalid
	}

	if err := s.requireMember(ctx, userID, teamID); err != nil {
		return nil, err
	}

	product := &Product{
		TeamID:      teamID,
		Name:        name,
		Description: description,
		CreatedBy:   userID,
	}
	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

func (s *Service) ListProducts(ctx context.Context, userID uint, teamID uint) ([]ProductResponse, error) {
	if userID == 0 || teamID == 0 {
		return nil, ErrForbidden
	}

	if err := s.requireMember(ctx, userID, teamID); err != nil {
		return nil, err
	}

	products, err := s.repo.ListByTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	response := make([]ProductResponse, 0, len(products))
	for _, product := range products {
		response = append(response, product.ToResponse())
	}
	return response, nil
}

func (s *Service) GetProduct(ctx context.Context, userID uint, productID uint) (*ProductResponse, error) {
	if userID == 0 || productID == 0 {
		return nil, ErrForbidden
	}

	product, err := s.repo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	if err := s.requireMember(ctx, userID, product.TeamID); err != nil {
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

func (s *Service) GetProductSummary(ctx context.Context, userID uint, productID uint) (*ProductSummaryResponse, error) {
	if userID == 0 || productID == 0 {
		return nil, ErrForbidden
	}
	product, err := s.repo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if err := s.requireMember(ctx, userID, product.TeamID); err != nil {
		return nil, err
	}
	stats, err := s.repo.SummaryStats(ctx, productID)
	if err != nil {
		return nil, err
	}
	return &ProductSummaryResponse{
		Product:             product.ToResponse(),
		ProductSummaryStats: *stats,
	}, nil
}

func (s *Service) requireMember(ctx context.Context, userID uint, teamID uint) error {
	ok, err := s.membership.IsMember(ctx, userID, teamID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}
