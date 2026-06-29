package requirement

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrRequirementNotFound = errors.New("requirement not found")
)

type RepositoryPort interface {
	Create(ctx context.Context, req *Requirement) error
	ListByProduct(ctx context.Context, productID uint, status string, page int, pageSize int) ([]Requirement, int64, error)
	FindByID(ctx context.Context, id uint) (*Requirement, error)
	Update(ctx context.Context, req *Requirement) error
	Delete(ctx context.Context, id uint) error
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, req *Requirement) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *Repository) ListByProduct(ctx context.Context, productID uint, status string, page int, pageSize int) ([]Requirement, int64, error) {
	db := r.db.WithContext(ctx).Model(&Requirement{}).Where("product_id = ?", productID)
	if status != "" {
		db = db.Where("status = ?", status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize)
	offset := (page - 1) * pageSize

	var items []Requirement
	err := db.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*Requirement, error) {
	var req Requirement
	err := r.db.WithContext(ctx).First(&req, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRequirementNotFound
	}
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *Repository) Update(ctx context.Context, req *Requirement) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&Requirement{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRequirementNotFound
	}
	return nil
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
