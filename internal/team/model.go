package team

import (
	"time"

	"pulseroad/internal/pkg/database"
)

const RoleOwner = "owner"

type Team struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"type:varchar(120);not null"`
	Description string    `json:"description" gorm:"type:varchar(500)"`
	CreatedBy   uint      `json:"created_by" gorm:"not null;index"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TeamMember struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TeamID    uint      `json:"team_id" gorm:"not null;uniqueIndex:idx_team_members_team_user"`
	UserID    uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_team_members_team_user;index"`
	Role      string    `json:"role" gorm:"type:varchar(32);not null"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamWithRole struct {
	Team Team
	Role string
}

type TeamResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   uint      `json:"created_by"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

func (t TeamWithRole) ToResponse() TeamResponse {
	return TeamResponse{
		ID:          t.Team.ID,
		Name:        t.Team.Name,
		Description: t.Team.Description,
		CreatedBy:   t.Team.CreatedBy,
		Role:        t.Role,
		CreatedAt:   t.Team.CreatedAt,
	}
}

func init() {
	database.RegisterModel(&Team{}, &TeamMember{})
}
