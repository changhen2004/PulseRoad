package product

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, product *Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *Repository) ListByTeam(ctx context.Context, teamID uint) ([]Product, error) {
	var products []Product
	err := r.db.WithContext(ctx).
		Where("team_id = ?", teamID).
		Order("id ASC").
		Find(&products).Error
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*Product, error) {
	var product Product
	err := r.db.WithContext(ctx).First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	return &product, nil
}
