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
	ID          uint      `json:"id" gorm:"primaryKey"`
	ProductID   uint      `json:"product_id" gorm:"not null;index"`
	Title       string    `json:"title" gorm:"type:varchar(160);not null"`
	Description string    `json:"description" gorm:"type:varchar(1000)"`
	Status      string    `json:"status" gorm:"type:varchar(32);not null;index"`
	CreatedBy   uint      `json:"created_by" gorm:"not null;index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FeedbackResponse struct {
	ID          uint      `json:"id"`
	ProductID   uint      `json:"product_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedBy   uint      `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

func (f Feedback) ToResponse() FeedbackResponse {
	return FeedbackResponse{
		ID:          f.ID,
		ProductID:   f.ProductID,
		Title:       f.Title,
		Description: f.Description,
		Status:      f.Status,
		CreatedBy:   f.CreatedBy,
		CreatedAt:   f.CreatedAt,
	}
}

func init() {
	database.RegisterModel(&Feedback{})
}
