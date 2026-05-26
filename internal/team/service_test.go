package team

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

type fakeTeamRepository struct {
	nextTeamID   uint
	nextMemberID uint
	teams        map[uint]*Team
	members      map[uint]map[uint]*TeamMember
}

func newFakeTeamRepository() *fakeTeamRepository {
	return &fakeTeamRepository{
		nextTeamID:   1,
		nextMemberID: 1,
		teams:        make(map[uint]*Team),
		members:      make(map[uint]map[uint]*TeamMember),
	}
}

func (r *fakeTeamRepository) CreateTeamWithOwner(_ context.Context, team *Team, member *TeamMember) error {
	team.ID = r.nextTeamID
	team.CreatedAt = time.Now()
	team.UpdatedAt = team.CreatedAt
	r.nextTeamID++

	member.ID = r.nextMemberID
	member.TeamID = team.ID
	member.CreatedAt = time.Now()
	r.nextMemberID++

	copyTeam := *team
	copyMember := *member
	r.teams[team.ID] = &copyTeam
	if r.members[team.ID] == nil {
		r.members[team.ID] = make(map[uint]*TeamMember)
	}
	r.members[team.ID][member.UserID] = &copyMember
	return nil
}

func (r *fakeTeamRepository) ListTeamsForUser(_ context.Context, userID uint) ([]TeamWithRole, error) {
	var teams []TeamWithRole
	for teamID, byUser := range r.members {
		member, ok := byUser[userID]
		if !ok {
			continue
		}
		team := r.teams[teamID]
		teams = append(teams, TeamWithRole{Team: *team, Role: member.Role})
	}
	return teams, nil
}

func (r *fakeTeamRepository) FindTeamForUser(_ context.Context, teamID uint, userID uint) (*TeamWithRole, error) {
	team, ok := r.teams[teamID]
	if !ok {
		return nil, ErrTeamNotFound
	}
	member, ok := r.members[teamID][userID]
	if !ok {
		return nil, ErrForbidden
	}
	return &TeamWithRole{Team: *team, Role: member.Role}, nil
}

func TestCreateTeamAddsCreatorAsOwner(t *testing.T) {
	svc := NewService(newFakeTeamRepository())

	created, err := svc.CreateTeam(context.Background(), 7, CreateTeamInput{
		Name:        "Core",
		Description: "Roadmap work",
	})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}

	if created.ID == 0 {
		t.Fatal("expected team id")
	}
	if created.Name != "Core" || created.Description != "Roadmap work" {
		t.Fatalf("unexpected team response: %#v", created)
	}
	if created.CreatedBy != 7 || created.Role != RoleOwner {
		t.Fatalf("expected creator 7 as owner, got created_by=%d role=%s", created.CreatedBy, created.Role)
	}
}

func TestListTeamsReturnsOnlyJoinedTeams(t *testing.T) {
	repo := newFakeTeamRepository()
	svc := NewService(repo)

	if _, err := svc.CreateTeam(context.Background(), 7, CreateTeamInput{Name: "Core"}); err != nil {
		t.Fatalf("create team for user 7: %v", err)
	}
	if _, err := svc.CreateTeam(context.Background(), 8, CreateTeamInput{Name: "Private"}); err != nil {
		t.Fatalf("create team for user 8: %v", err)
	}

	teams, err := svc.ListTeams(context.Background(), 7)
	if err != nil {
		t.Fatalf("list teams: %v", err)
	}

	if len(teams) != 1 {
		t.Fatalf("expected one joined team, got %#v", teams)
	}
	if teams[0].Name != "Core" || teams[0].Role != RoleOwner {
		t.Fatalf("unexpected team list: %#v", teams)
	}
}

func TestGetTeamRejectsNonMember(t *testing.T) {
	svc := NewService(newFakeTeamRepository())
	created, err := svc.CreateTeam(context.Background(), 7, CreateTeamInput{Name: "Core"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}

	_, err = svc.GetTeam(context.Background(), 8, created.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestTeamMemberHasUniqueTeamUserConstraint(t *testing.T) {
	modelType := reflect.TypeOf(TeamMember{})

	teamID, ok := modelType.FieldByName("TeamID")
	if !ok {
		t.Fatal("TeamMember.TeamID missing")
	}
	userID, ok := modelType.FieldByName("UserID")
	if !ok {
		t.Fatal("TeamMember.UserID missing")
	}

	teamTag := teamID.Tag.Get("gorm")
	userTag := userID.Tag.Get("gorm")
	if !strings.Contains(teamTag, "uniqueIndex:idx_team_members_team_user") {
		t.Fatalf("TeamID missing shared unique index: %q", teamTag)
	}
	if !strings.Contains(userTag, "uniqueIndex:idx_team_members_team_user") {
		t.Fatalf("UserID missing shared unique index: %q", userTag)
	}
}
