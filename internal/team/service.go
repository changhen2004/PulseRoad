package team

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrForbidden          = errors.New("forbidden")
	ErrInvalid            = errors.New("invalid input")
	ErrInvitationExists   = errors.New("invitation already exists")
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrLastOwner          = errors.New("team must keep at least one owner")
	ErrMemberNotFound     = errors.New("member not found")
	ErrTeamNotFound       = errors.New("team not found")
	ErrUserNotFound       = errors.New("user not found")
)

type CreateTeamInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type InviteMemberInput struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateMemberRoleInput struct {
	Role string `json:"role"`
}

type TeamRepository interface {
	CreateTeamWithOwner(ctx context.Context, team *Team, member *TeamMember) error
	ListTeamsForUser(ctx context.Context, userID uint) ([]TeamWithRole, error)
	FindTeamForUser(ctx context.Context, teamID uint, userID uint) (*TeamWithRole, error)
	FindUserByID(ctx context.Context, userID uint) (*UserBrief, error)
	FindUserByEmail(ctx context.Context, email string) (*UserBrief, error)
	CreateInvitation(ctx context.Context, invitation *TeamInvitation) error
	ListInvitationsForEmail(ctx context.Context, email string) ([]TeamInvitationWithTeam, error)
	FindInvitationByID(ctx context.Context, invitationID uint) (*TeamInvitation, error)
	AcceptInvitation(ctx context.Context, invitation *TeamInvitation, member *TeamMember) error
	ListMembers(ctx context.Context, teamID uint) ([]TeamMemberResponse, error)
	FindMember(ctx context.Context, teamID uint, userID uint) (*TeamMember, error)
	UpdateMemberRole(ctx context.Context, teamID uint, userID uint, role string) error
	DeleteMember(ctx context.Context, teamID uint, userID uint) error
	CountOwners(ctx context.Context, teamID uint) (int64, error)
}

type Service struct {
	repo TeamRepository
}

func NewService(repo TeamRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTeam(ctx context.Context, userID uint, input CreateTeamInput) (*TeamResponse, error) {
	name := strings.TrimSpace(input.Name)
	description := strings.TrimSpace(input.Description)
	if userID == 0 || name == "" {
		return nil, ErrInvalid
	}

	team := &Team{
		Name:        name,
		Description: description,
		CreatedBy:   userID,
	}
	member := &TeamMember{
		UserID: userID,
		Role:   RoleOwner,
	}

	if err := s.repo.CreateTeamWithOwner(ctx, team, member); err != nil {
		return nil, err
	}

	response := TeamWithRole{Team: *team, Role: member.Role}.ToResponse()
	return &response, nil
}

func (s *Service) ListTeams(ctx context.Context, userID uint) ([]TeamResponse, error) {
	if userID == 0 {
		return nil, ErrForbidden
	}

	teams, err := s.repo.ListTeamsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]TeamResponse, 0, len(teams))
	for _, team := range teams {
		response = append(response, team.ToResponse())
	}
	return response, nil
}

func (s *Service) GetTeam(ctx context.Context, userID uint, teamID uint) (*TeamResponse, error) {
	if userID == 0 || teamID == 0 {
		return nil, ErrForbidden
	}

	team, err := s.repo.FindTeamForUser(ctx, teamID, userID)
	if err != nil {
		return nil, err
	}

	response := team.ToResponse()
	return &response, nil
}

func (s *Service) IsMember(ctx context.Context, userID uint, teamID uint) (bool, error) {
	if userID == 0 || teamID == 0 {
		return false, nil
	}

	_, err := s.repo.FindTeamForUser(ctx, teamID, userID)
	if errors.Is(err, ErrForbidden) || errors.Is(err, ErrTeamNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Service) InviteMember(ctx context.Context, userID uint, teamID uint, input InviteMemberInput) (*TeamInvitationResponse, error) {
	email := normalizeEmail(input.Email)
	role := normalizeRole(input.Role)
	if userID == 0 || teamID == 0 || email == "" || !validRole(role) {
		return nil, ErrInvalid
	}

	team, err := s.requireOwner(ctx, userID, teamID)
	if err != nil {
		return nil, err
	}
	user, err := s.repo.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.FindMember(ctx, teamID, user.ID); err == nil {
		return nil, ErrInvitationExists
	} else if !errors.Is(err, ErrMemberNotFound) {
		return nil, err
	}

	invitation := &TeamInvitation{
		TeamID:    teamID,
		Email:     email,
		Role:      role,
		Status:    InvitationPending,
		InvitedBy: userID,
	}
	if err := s.repo.CreateInvitation(ctx, invitation); err != nil {
		return nil, err
	}

	response := invitation.ToResponse(team.Team.Name)
	return &response, nil
}

func (s *Service) ListMembers(ctx context.Context, userID uint, teamID uint) ([]TeamMemberResponse, error) {
	if userID == 0 || teamID == 0 {
		return nil, ErrForbidden
	}
	if _, err := s.repo.FindTeamForUser(ctx, teamID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(ctx, teamID)
}

func (s *Service) ListInvitations(ctx context.Context, userID uint) ([]TeamInvitationResponse, error) {
	if userID == 0 {
		return nil, ErrForbidden
	}
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	invitations, err := s.repo.ListInvitationsForEmail(ctx, user.Email)
	if err != nil {
		return nil, err
	}
	response := make([]TeamInvitationResponse, 0, len(invitations))
	for _, invitation := range invitations {
		response = append(response, invitation.ToResponse())
	}
	return response, nil
}

func (s *Service) AcceptInvitationForUser(ctx context.Context, userID uint, invitationID uint) (*TeamInvitationResponse, error) {
	if userID == 0 || invitationID == 0 {
		return nil, ErrForbidden
	}
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.AcceptInvitation(ctx, userID, user.Email, invitationID)
}

func (s *Service) AcceptInvitation(ctx context.Context, userID uint, email string, invitationID uint) (*TeamInvitationResponse, error) {
	email = normalizeEmail(email)
	if userID == 0 || email == "" || invitationID == 0 {
		return nil, ErrInvalid
	}
	invitation, err := s.repo.FindInvitationByID(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	if invitation.Status != InvitationPending || invitation.Email != email {
		return nil, ErrInvitationNotFound
	}
	if _, err := s.repo.FindMember(ctx, invitation.TeamID, userID); err == nil {
		return nil, ErrInvitationExists
	} else if !errors.Is(err, ErrMemberNotFound) {
		return nil, err
	}

	member := &TeamMember{
		TeamID: invitation.TeamID,
		UserID: userID,
		Role:   invitation.Role,
	}
	if err := s.repo.AcceptInvitation(ctx, invitation, member); err != nil {
		return nil, err
	}

	team, err := s.repo.FindTeamForUser(ctx, invitation.TeamID, userID)
	if err != nil {
		return nil, err
	}
	response := invitation.ToResponse(team.Team.Name)
	return &response, nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, userID uint, teamID uint, memberUserID uint, input UpdateMemberRoleInput) (*TeamMemberResponse, error) {
	role := normalizeRole(input.Role)
	if userID == 0 || teamID == 0 || memberUserID == 0 || !validRole(role) {
		return nil, ErrInvalid
	}
	if _, err := s.requireOwner(ctx, userID, teamID); err != nil {
		return nil, err
	}
	member, err := s.repo.FindMember(ctx, teamID, memberUserID)
	if err != nil {
		return nil, err
	}
	if member.Role == RoleOwner && role != RoleOwner {
		if err := s.ensureAnotherOwner(ctx, teamID); err != nil {
			return nil, err
		}
	}
	if err := s.repo.UpdateMemberRole(ctx, teamID, memberUserID, role); err != nil {
		return nil, err
	}
	members, err := s.repo.ListMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}
	for _, member := range members {
		if member.UserID == memberUserID {
			return &member, nil
		}
	}
	return nil, ErrMemberNotFound
}

func (s *Service) RemoveMember(ctx context.Context, userID uint, teamID uint, memberUserID uint) error {
	if userID == 0 || teamID == 0 || memberUserID == 0 {
		return ErrInvalid
	}
	if _, err := s.requireOwner(ctx, userID, teamID); err != nil {
		return err
	}
	member, err := s.repo.FindMember(ctx, teamID, memberUserID)
	if err != nil {
		return err
	}
	if member.Role == RoleOwner {
		if err := s.ensureAnotherOwner(ctx, teamID); err != nil {
			return err
		}
	}
	return s.repo.DeleteMember(ctx, teamID, memberUserID)
}

func (s *Service) requireOwner(ctx context.Context, userID uint, teamID uint) (*TeamWithRole, error) {
	team, err := s.repo.FindTeamForUser(ctx, teamID, userID)
	if err != nil {
		return nil, err
	}
	if team.Role != RoleOwner {
		return nil, ErrForbidden
	}
	return team, nil
}

func (s *Service) ensureAnotherOwner(ctx context.Context, teamID uint) error {
	count, err := s.repo.CountOwners(ctx, teamID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrLastOwner
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeRole(role string) string {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		return RoleMember
	}
	return role
}

func validRole(role string) bool {
	return role == RoleOwner || role == RoleMember
}
