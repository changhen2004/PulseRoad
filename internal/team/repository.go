package team

import (
	"context"
	"errors"

	"pulseroad/internal/auth"

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

func (r *Repository) FindUserByID(ctx context.Context, userID uint) (*UserBrief, error) {
	var user auth.User
	err := r.db.WithContext(ctx).First(&user, userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &UserBrief{ID: user.ID, Email: user.Email, Name: user.Name}, nil
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*UserBrief, error) {
	var user auth.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &UserBrief{ID: user.ID, Email: user.Email, Name: user.Name}, nil
}

func (r *Repository) CreateInvitation(ctx context.Context, invitation *TeamInvitation) error {
	var existing TeamInvitation
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND email = ? AND status = ?", invitation.TeamID, invitation.Email, InvitationPending).
		First(&existing).Error
	if err == nil {
		return ErrInvitationExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return r.db.WithContext(ctx).Create(invitation).Error
}

func (r *Repository) ListInvitationsForEmail(ctx context.Context, email string) ([]TeamInvitationWithTeam, error) {
	var rows []struct {
		TeamInvitation
		TeamName string
	}
	err := r.db.WithContext(ctx).
		Table("team_invitations").
		Select("team_invitations.*, teams.name AS team_name").
		Joins("JOIN teams ON teams.id = team_invitations.team_id").
		Where("team_invitations.email = ? AND team_invitations.status = ?", email, InvitationPending).
		Order("team_invitations.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	invitations := make([]TeamInvitationWithTeam, 0, len(rows))
	for _, row := range rows {
		invitations = append(invitations, TeamInvitationWithTeam{
			Invitation: row.TeamInvitation,
			Team:       Team{ID: row.TeamID, Name: row.TeamName},
		})
	}
	return invitations, nil
}

func (r *Repository) FindInvitationByID(ctx context.Context, invitationID uint) (*TeamInvitation, error) {
	var invitation TeamInvitation
	err := r.db.WithContext(ctx).First(&invitation, invitationID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvitationNotFound
	}
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *Repository) AcceptInvitation(ctx context.Context, invitation *TeamInvitation, member *TeamMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(member).Error; err != nil {
			return err
		}
		if err := tx.Model(&TeamInvitation{}).
			Where("id = ? AND status = ?", invitation.ID, InvitationPending).
			Updates(map[string]any{
				"status":      InvitationAccepted,
				"accepted_at": gorm.Expr("CURRENT_TIMESTAMP"),
			}).Error; err != nil {
			return err
		}
		return tx.First(invitation, invitation.ID).Error
	})
}

func (r *Repository) ListMembers(ctx context.Context, teamID uint) ([]TeamMemberResponse, error) {
	var members []TeamMemberResponse
	err := r.db.WithContext(ctx).
		Table("team_members").
		Select("team_members.user_id, users.email, users.name, team_members.role, team_members.created_at").
		Joins("JOIN users ON users.id = team_members.user_id").
		Where("team_members.team_id = ?", teamID).
		Order("team_members.id ASC").
		Scan(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (r *Repository) FindMember(ctx context.Context, teamID uint, userID uint) (*TeamMember, error) {
	var member TeamMember
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *Repository) UpdateMemberRole(ctx context.Context, teamID uint, userID uint, role string) error {
	result := r.db.WithContext(ctx).
		Model(&TeamMember{}).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Update("role", role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (r *Repository) DeleteMember(ctx context.Context, teamID uint, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Delete(&TeamMember{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return nil
}

func (r *Repository) CountOwners(ctx context.Context, teamID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&TeamMember{}).
		Where("team_id = ? AND role = ?", teamID, RoleOwner).
		Count(&count).Error
	return count, err
}
