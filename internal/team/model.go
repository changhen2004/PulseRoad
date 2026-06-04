package team

import (
	"time"

	"pulseroad/internal/pkg/database"
)

const RoleOwner = "owner"
const RoleMember = "member"

const (
	InvitationPending  = "pending"
	InvitationAccepted = "accepted"
)

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

type TeamInvitation struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	TeamID     uint       `json:"team_id" gorm:"not null;index"`
	Email      string     `json:"email" gorm:"type:varchar(255);not null;index"`
	Role       string     `json:"role" gorm:"type:varchar(32);not null"`
	Status     string     `json:"status" gorm:"type:varchar(32);not null;index"`
	InvitedBy  uint       `json:"invited_by" gorm:"not null;index"`
	AcceptedAt *time.Time `json:"accepted_at"`
	CreatedAt  time.Time  `json:"created_at"`
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

type UserBrief struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type TeamMemberResponse struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamInvitationWithTeam struct {
	Invitation TeamInvitation
	Team       Team
}

type TeamInvitationResponse struct {
	ID         uint       `json:"id"`
	TeamID     uint       `json:"team_id"`
	TeamName   string     `json:"team_name"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	Status     string     `json:"status"`
	InvitedBy  uint       `json:"invited_by"`
	AcceptedAt *time.Time `json:"accepted_at"`
	CreatedAt  time.Time  `json:"created_at"`
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

func (i TeamInvitationWithTeam) ToResponse() TeamInvitationResponse {
	return TeamInvitationResponse{
		ID:         i.Invitation.ID,
		TeamID:     i.Invitation.TeamID,
		TeamName:   i.Team.Name,
		Email:      i.Invitation.Email,
		Role:       i.Invitation.Role,
		Status:     i.Invitation.Status,
		InvitedBy:  i.Invitation.InvitedBy,
		AcceptedAt: i.Invitation.AcceptedAt,
		CreatedAt:  i.Invitation.CreatedAt,
	}
}

func (i TeamInvitation) ToResponse(teamName string) TeamInvitationResponse {
	return TeamInvitationResponse{
		ID:         i.ID,
		TeamID:     i.TeamID,
		TeamName:   teamName,
		Email:      i.Email,
		Role:       i.Role,
		Status:     i.Status,
		InvitedBy:  i.InvitedBy,
		AcceptedAt: i.AcceptedAt,
		CreatedAt:  i.CreatedAt,
	}
}

func init() {
	database.RegisterModel(&Team{}, &TeamMember{}, &TeamInvitation{})
}
