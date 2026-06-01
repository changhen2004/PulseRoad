package flagflow

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

func (r *Repository) Create(ctx context.Context, flag *FeatureFlag) error {
	return r.db.WithContext(ctx).Create(flag).Error
}

func (r *Repository) ListByProduct(ctx context.Context, productID uint, environment string) ([]FeatureFlag, error) {
	var flags []FeatureFlag
	query := r.db.WithContext(ctx).Where("product_id = ?", productID)
	if environment != "" {
		query = query.Where("environment = ?", environment)
	}
	err := query.Order("id ASC").Find(&flags).Error
	if err != nil {
		return nil, err
	}
	return flags, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*FeatureFlag, error) {
	var flag FeatureFlag
	err := r.db.WithContext(ctx).First(&flag, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFlagNotFound
	}
	if err != nil {
		return nil, err
	}
	return &flag, nil
}

func (r *Repository) FindByKey(ctx context.Context, productID uint, environment string, key string) (*FeatureFlag, error) {
	var flag FeatureFlag
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND environment = ? AND `key` = ?", productID, environment, key).
		First(&flag).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFlagNotFound
	}
	if err != nil {
		return nil, err
	}
	return &flag, nil
}

func (r *Repository) Update(ctx context.Context, flag *FeatureFlag) error {
	return r.db.WithContext(ctx).Save(flag).Error
}
