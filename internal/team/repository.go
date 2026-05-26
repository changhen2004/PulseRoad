package team

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

func (r *Repository) CreateTeamWithOwner(ctx context.Context, team *Team, member *TeamMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(team).Error; err != nil {
			return err
		}
		member.TeamID = team.ID
		return tx.Create(member).Error
	})
}

func (r *Repository) ListTeamsForUser(ctx context.Context, userID uint) ([]TeamWithRole, error) {
	var rows []struct {
		Team
		Role string
	}

	err := r.db.WithContext(ctx).
		Table("teams").
		Select("teams.*, team_members.role").
		Joins("JOIN team_members ON team_members.team_id = teams.id").
		Where("team_members.user_id = ?", userID).
		Order("teams.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	teams := make([]TeamWithRole, 0, len(rows))
	for _, row := range rows {
		teams = append(teams, TeamWithRole{Team: row.Team, Role: row.Role})
	}
	return teams, nil
}

func (r *Repository) FindTeamForUser(ctx context.Context, teamID uint, userID uint) (*TeamWithRole, error) {
	var team Team
	err := r.db.WithContext(ctx).First(&team, teamID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTeamNotFound
	}
	if err != nil {
		return nil, err
	}

	var member TeamMember
	err = r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrForbidden
	}
	if err != nil {
		return nil, err
	}

	return &TeamWithRole{Team: team, Role: member.Role}, nil
}
