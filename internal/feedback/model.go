package feedback

import (
	"time"

	"pulseroad/internal/pkg/database"
)

const (
	StatusOpen     = "open"
	StatusResolved = "resolved"
)

type Feedback struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ProductID    uint      `json:"product_id" gorm:"not null;index"`
	Title        string    `json:"title" gorm:"type:varchar(160);not null"`
	Content      string    `json:"content" gorm:"type:text;not null"`
	Status       string    `json:"status" gorm:"type:varchar(32);not null;index"`
	CreatedBy    uint      `json:"created_by" gorm:"not null;index"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	VoteCount    int64     `json:"-" gorm:"column:vote_count;->;-:migration"`
	CommentCount int64     `json:"-" gorm:"column:comment_count;->;-:migration"`
	Voted        bool      `json:"-" gorm:"column:voted;->;-:migration"`
}

type FeedbackResponse struct {
	ID           uint      `json:"id"`
	ProductID    uint      `json:"product_id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Status       string    `json:"status"`
	CreatedBy    uint      `json:"created_by"`
	VoteCount    int64     `json:"vote_count"`
	CommentCount int64     `json:"comment_count"`
	Voted        bool      `json:"voted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type FeedbackComment struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	FeedbackID uint      `json:"feedback_id" gorm:"not null;index"`
	Content    string    `json:"content" gorm:"type:text;not null"`
	CreatedBy  uint      `json:"created_by" gorm:"not null;index"`
	CreatedAt  time.Time `json:"created_at"`
}

type FeedbackCommentResponse struct {
	ID         uint      `json:"id"`
	FeedbackID uint      `json:"feedback_id"`
	Content    string    `json:"content"`
	CreatedBy  uint      `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type FeedbackVote struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	FeedbackID uint      `json:"feedback_id" gorm:"not null;uniqueIndex:idx_feedback_votes_feedback_user"`
	UserID     uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_feedback_votes_feedback_user;index"`
	CreatedAt  time.Time `json:"created_at"`
}

type FeedbackPageResponse struct {
	Items    []FeedbackResponse `json:"items"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int64              `json:"total"`
}

type FeedbackVoteResponse struct {
	FeedbackID uint  `json:"feedback_id"`
	Voted      bool  `json:"voted"`
	VoteCount  int64 `json:"vote_count"`
}

func (f Feedback) ToResponse() FeedbackResponse {
	return FeedbackResponse{
		ID:           f.ID,
		ProductID:    f.ProductID,
		Title:        f.Title,
		Content:      f.Content,
		Status:       f.Status,
		CreatedBy:    f.CreatedBy,
		VoteCount:    f.VoteCount,
		CommentCount: f.CommentCount,
		Voted:        f.Voted,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}

func (c FeedbackComment) ToResponse() FeedbackCommentResponse {
	return FeedbackCommentResponse{
		ID:         c.ID,
		FeedbackID: c.FeedbackID,
		Content:    c.Content,
		CreatedBy:  c.CreatedBy,
		CreatedAt:  c.CreatedAt,
	}
}

func init() {
	database.RegisterModel(&Feedback{}, &FeedbackComment{}, &FeedbackVote{})
}
