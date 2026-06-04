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

func (r *Repository) SummaryStats(ctx context.Context, productID uint) (*ProductSummaryStats, error) {
	db := r.db.WithContext(ctx)
	var stats ProductSummaryStats
	if err := db.Table("feedbacks").Where("product_id = ?", productID).Count(&stats.FeedbackTotal).Error; err != nil {
		return nil, err
	}
	if err := db.Table("feedbacks").Where("product_id = ? AND status = ?", productID, "open").Count(&stats.FeedbackOpen).Error; err != nil {
		return nil, err
	}
	if err := db.Table("feedbacks").Where("product_id = ? AND status = ?", productID, "resolved").Count(&stats.FeedbackResolved).Error; err != nil {
		return nil, err
	}
	if err := db.Table("feedback_comments").
		Joins("JOIN feedbacks ON feedbacks.id = feedback_comments.feedback_id").
		Where("feedbacks.product_id = ?", productID).
		Count(&stats.CommentTotal).Error; err != nil {
		return nil, err
	}
	if err := db.Table("feedback_votes").
		Joins("JOIN feedbacks ON feedbacks.id = feedback_votes.feedback_id").
		Where("feedbacks.product_id = ?", productID).
		Count(&stats.VoteTotal).Error; err != nil {
		return nil, err
	}
	if err := db.Table("feature_flags").Where("product_id = ?", productID).Count(&stats.FlagTotal).Error; err != nil {
		return nil, err
	}
	if err := db.Table("feature_flags").Where("product_id = ? AND enabled = ?", productID, true).Count(&stats.FlagEnabled).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}
