package feedback

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

func (r *Repository) Create(ctx context.Context, feedback *Feedback) error {
	return r.db.WithContext(ctx).Create(feedback).Error
}

func (r *Repository) ListByProduct(ctx context.Context, productID uint) ([]Feedback, error) {
	var feedbackItems []Feedback
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("created_at DESC").
		Find(&feedbackItems).Error
	if err != nil {
		return nil, err
	}
	return feedbackItems, nil
}

func (r *Repository) FindByID(ctx context.Context, id uint) (*Feedback, error) {
	var feedback Feedback
	err := r.db.WithContext(ctx).First(&feedback, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFeedbackNotFound
	}
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id uint, status string) (*Feedback, error) {
	db := r.db.WithContext(ctx)
	if err := db.
		Model(&Feedback{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return nil, err
	}

	var feedback Feedback
	err := db.First(&feedback, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFeedbackNotFound
	}
	if err != nil {
		return nil, err
	}
	return &feedback, nil
}
