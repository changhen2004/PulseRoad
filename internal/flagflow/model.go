package flagflow

import (
	"time"

	"pulseroad/internal/pkg/database"
)

const (
	DefaultEnvironment = "development"

	EventActionCreated = "created"
	EventActionUpdated = "updated"
	EventActionToggled = "toggled"
)

type FeatureFlag struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	ProductID         uint      `json:"product_id" gorm:"not null;uniqueIndex:idx_flags_product_env_key"`
	Key               string    `json:"key" gorm:"type:varchar(120);not null;uniqueIndex:idx_flags_product_env_key"`
	Name              string    `json:"name" gorm:"type:varchar(160);not null"`
	Description       string    `json:"description" gorm:"type:text"`
	Environment       string    `json:"environment" gorm:"type:varchar(40);not null;uniqueIndex:idx_flags_product_env_key"`
	Enabled           bool      `json:"enabled" gorm:"not null;default:false"`
	RolloutPercentage int       `json:"rollout_percentage" gorm:"not null;default:0"`
	CreatedBy         uint      `json:"created_by" gorm:"not null;index"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type FeatureFlagResponse struct {
	ID                uint      `json:"id"`
	ProductID         uint      `json:"product_id"`
	Key               string    `json:"key"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Environment       string    `json:"environment"`
	Enabled           bool      `json:"enabled"`
	RolloutPercentage int       `json:"rollout_percentage"`
	CreatedBy         uint      `json:"created_by"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (f FeatureFlag) ToResponse() FeatureFlagResponse {
	return FeatureFlagResponse{
		ID:                f.ID,
		ProductID:         f.ProductID,
		Key:               f.Key,
		Name:              f.Name,
		Description:       f.Description,
		Environment:       f.Environment,
		Enabled:           f.Enabled,
		RolloutPercentage: f.RolloutPercentage,
		CreatedBy:         f.CreatedBy,
		CreatedAt:         f.CreatedAt,
		UpdatedAt:         f.UpdatedAt,
	}
}

func init() {
	database.RegisterModel(&FeatureFlag{})
}
