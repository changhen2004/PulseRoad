package product

import (
	"time"

	"pulseroad/internal/pkg/database"
)

type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TeamID      uint      `json:"team_id" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"type:varchar(120);not null"`
	Description string    `json:"description" gorm:"type:varchar(500)"`
	CreatedBy   uint      `json:"created_by" gorm:"not null;index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductResponse struct {
	ID          uint      `json:"id"`
	TeamID      uint      `json:"team_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProductSummaryStats struct {
	FeedbackTotal    int64 `json:"feedback_total"`
	FeedbackOpen     int64 `json:"feedback_open"`
	FeedbackResolved int64 `json:"feedback_resolved"`
	CommentTotal     int64 `json:"comment_total"`
	VoteTotal        int64 `json:"vote_total"`
	FlagTotal        int64 `json:"flag_total"`
	FlagEnabled      int64 `json:"flag_enabled"`
}

type ProductSummaryResponse struct {
	Product ProductResponse `json:"product"`
	ProductSummaryStats
}

func (p Product) ToResponse() ProductResponse {
	return ProductResponse{
		ID:          p.ID,
		TeamID:      p.TeamID,
		Name:        p.Name,
		Description: p.Description,
		CreatedBy:   p.CreatedBy,
		CreatedAt:   p.CreatedAt,
	}
}

func init() {
	database.RegisterModel(&Product{})
}
