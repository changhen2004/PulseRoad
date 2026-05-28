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
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProductID uint      `json:"product_id" gorm:"not null;index"`
	Title     string    `json:"title" gorm:"type:varchar(160);not null"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Status    string    `json:"status" gorm:"type:varchar(32);not null;index"`
	CreatedBy uint      `json:"created_by" gorm:"not null;index"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FeedbackResponse struct {
	ID        uint      `json:"id"`
	ProductID uint      `json:"product_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedBy uint      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (f Feedback) ToResponse() FeedbackResponse {
	return FeedbackResponse{
		ID:        f.ID,
		ProductID: f.ProductID,
		Title:     f.Title,
		Content:   f.Content,
		Status:    f.Status,
		CreatedBy: f.CreatedBy,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}
}

func init() {
	database.RegisterModel(&Feedback{})
}
