package requirement

import (
	"time"

	"pulseroad/internal/pkg/database"
)

const (
	StatusPlanned    = "planned"
	StatusInProgress = "in_progress"
	StatusReleased   = "released"

	PriorityP0 = "p0"
	PriorityP1 = "p1"
	PriorityP2 = "p2"
	PriorityP3 = "p3"
)

type Requirement struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	ProductID         uint      `json:"product_id" gorm:"not null;index"`
	Title             string    `json:"title" gorm:"type:varchar(160);not null"`
	Description       string    `json:"description" gorm:"type:text"`
	Status            string    `json:"status" gorm:"type:varchar(32);not null;index"`
	Priority          string    `json:"priority" gorm:"type:varchar(8);not null"`
	SourceFeedbackID  *uint     `json:"source_feedback_id" gorm:"index"`
	CreatedBy         uint      `json:"created_by" gorm:"not null;index"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type RequirementResponse struct {
	ID               uint      `json:"id"`
	ProductID        uint      `json:"product_id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Status           string    `json:"status"`
	Priority         string    `json:"priority"`
	SourceFeedbackID *uint     `json:"source_feedback_id"`
	CreatedBy        uint      `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type RequirementPageResponse struct {
	Items    []RequirementResponse `json:"items"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	Total    int64                 `json:"total"`
}

func (r Requirement) ToResponse() RequirementResponse {
	return RequirementResponse{
		ID:               r.ID,
		ProductID:        r.ProductID,
		Title:            r.Title,
		Description:      r.Description,
		Status:           r.Status,
		Priority:         r.Priority,
		SourceFeedbackID: r.SourceFeedbackID,
		CreatedBy:        r.CreatedBy,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

func validStatus(status string) bool {
	return status == StatusPlanned || status == StatusInProgress || status == StatusReleased
}

func validPriority(priority string) bool {
	return priority == PriorityP0 || priority == PriorityP1 || priority == PriorityP2 || priority == PriorityP3
}

func init() {
	database.RegisterModel(&Requirement{})
}
