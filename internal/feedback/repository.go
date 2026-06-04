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

func (r *Repository) ListByProductPage(ctx context.Context, productID uint, query ListFeedbackQuery) (FeedbackPage, error) {
	db := r.db.WithContext(ctx).Model(&Feedback{}).Where("product_id = ?", productID)
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return FeedbackPage{}, err
	}

	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	offset := (page - 1) * pageSize
	var feedbackItems []Feedback
	err := db.
		Select(`feedbacks.*, 
			(SELECT COUNT(*) FROM feedback_votes WHERE feedback_votes.feedback_id = feedbacks.id) AS vote_count,
			(SELECT COUNT(*) FROM feedback_comments WHERE feedback_comments.feedback_id = feedbacks.id) AS comment_count,
			EXISTS(SELECT 1 FROM feedback_votes WHERE feedback_votes.feedback_id = feedbacks.id AND feedback_votes.user_id = ?) AS voted`, query.UserID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&feedbackItems).Error
	if err != nil {
		return FeedbackPage{}, err
	}

	return FeedbackPage{Items: feedbackItems, Page: page, PageSize: pageSize, Total: total}, nil
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

func (r *Repository) CreateComment(ctx context.Context, comment *FeedbackComment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *Repository) ListComments(ctx context.Context, feedbackID uint) ([]FeedbackComment, error) {
	var comments []FeedbackComment
	err := r.db.WithContext(ctx).
		Where("feedback_id = ?", feedbackID).
		Order("id ASC").
		Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *Repository) CreateVote(ctx context.Context, vote *FeedbackVote) error {
	err := r.db.WithContext(ctx).Create(vote).Error
	if err != nil {
		return ErrVoteExists
	}
	return nil
}

func (r *Repository) DeleteVote(ctx context.Context, feedbackID uint, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("feedback_id = ? AND user_id = ?", feedbackID, userID).
		Delete(&FeedbackVote{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrVoteNotFound
	}
	return nil
}

func (r *Repository) CountVotes(ctx context.Context, feedbackID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&FeedbackVote{}).
		Where("feedback_id = ?", feedbackID).
		Count(&count).Error
	return count, err
}
