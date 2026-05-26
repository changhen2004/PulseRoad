package team

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrForbidden    = errors.New("forbidden")
	ErrInvalid      = errors.New("invalid input")
	ErrTeamNotFound = errors.New("team not found")
)

type CreateTeamInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TeamRepository interface {
	CreateTeamWithOwner(ctx context.Context, team *Team, member *TeamMember) error
	ListTeamsForUser(ctx context.Context, userID uint) ([]TeamWithRole, error)
	FindTeamForUser(ctx context.Context, teamID uint, userID uint) (*TeamWithRole, error)
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
