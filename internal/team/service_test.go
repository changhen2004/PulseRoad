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
	nextTeamID       uint
	nextMemberID     uint
	nextInvitationID uint
	teams            map[uint]*Team
	members          map[uint]map[uint]*TeamMember
	invitations      map[uint]*TeamInvitation
	usersByEmail     map[string]UserBrief
}

func newFakeTeamRepository() *fakeTeamRepository {
	return &fakeTeamRepository{
		nextTeamID:       1,
		nextMemberID:     1,
		nextInvitationID: 1,
		teams:            make(map[uint]*Team),
		members:          make(map[uint]map[uint]*TeamMember),
		invitations:      make(map[uint]*TeamInvitation),
		usersByEmail:     make(map[string]UserBrief),
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

func (r *fakeTeamRepository) FindUserByEmail(_ context.Context, email string) (*UserBrief, error) {
	user, ok := r.usersByEmail[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return &user, nil
}

func (r *fakeTeamRepository) FindUserByID(_ context.Context, userID uint) (*UserBrief, error) {
	for _, user := range r.usersByEmail {
		if user.ID == userID {
			return &user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (r *fakeTeamRepository) CreateInvitation(_ context.Context, invitation *TeamInvitation) error {
	for _, existing := range r.invitations {
		if existing.TeamID == invitation.TeamID && existing.Email == invitation.Email && existing.Status == InvitationPending {
			return ErrInvitationExists
		}
	}
	invitation.ID = r.nextInvitationID
	invitation.CreatedAt = time.Now()
	r.nextInvitationID++
	copyInvitation := *invitation
	r.invitations[invitation.ID] = &copyInvitation
	return nil
}

func (r *fakeTeamRepository) ListInvitationsForEmail(_ context.Context, email string) ([]TeamInvitationWithTeam, error) {
	var invitations []TeamInvitationWithTeam
	for _, invitation := range r.invitations {
		if invitation.Email == email && invitation.Status == InvitationPending {
			invitations = append(invitations, TeamInvitationWithTeam{Invitation: *invitation, Team: *r.teams[invitation.TeamID]})
		}
	}
	return invitations, nil
}

func (r *fakeTeamRepository) FindInvitationByID(_ context.Context, invitationID uint) (*TeamInvitation, error) {
	invitation, ok := r.invitations[invitationID]
	if !ok {
		return nil, ErrInvitationNotFound
	}
	copyInvitation := *invitation
	return &copyInvitation, nil
}

func (r *fakeTeamRepository) AcceptInvitation(_ context.Context, invitation *TeamInvitation, member *TeamMember) error {
	stored, ok := r.invitations[invitation.ID]
	if !ok {
		return ErrInvitationNotFound
	}
	stored.Status = InvitationAccepted
	now := time.Now()
	stored.AcceptedAt = &now
	invitation.Status = stored.Status
	invitation.AcceptedAt = stored.AcceptedAt

	member.ID = r.nextMemberID
	member.CreatedAt = now
	r.nextMemberID++
	if r.members[member.TeamID] == nil {
		r.members[member.TeamID] = make(map[uint]*TeamMember)
	}
	copyMember := *member
	r.members[member.TeamID][member.UserID] = &copyMember
	return nil
}

func (r *fakeTeamRepository) ListMembers(_ context.Context, teamID uint) ([]TeamMemberResponse, error) {
	var members []TeamMemberResponse
	for _, member := range r.members[teamID] {
		email := ""
		name := ""
		for _, user := range r.usersByEmail {
			if user.ID == member.UserID {
				email = user.Email
				name = user.Name
			}
		}
		members = append(members, TeamMemberResponse{
			UserID:    member.UserID,
			Email:     email,
			Name:      name,
			Role:      member.Role,
			CreatedAt: member.CreatedAt,
		})
	}
	return members, nil
}

func (r *fakeTeamRepository) FindMember(_ context.Context, teamID uint, userID uint) (*TeamMember, error) {
	member, ok := r.members[teamID][userID]
	if !ok {
		return nil, ErrMemberNotFound
	}
	copyMember := *member
	return &copyMember, nil
}

func (r *fakeTeamRepository) UpdateMemberRole(_ context.Context, teamID uint, userID uint, role string) error {
	member, ok := r.members[teamID][userID]
	if !ok {
		return ErrMemberNotFound
	}
	member.Role = role
	return nil
}

func (r *fakeTeamRepository) DeleteMember(_ context.Context, teamID uint, userID uint) error {
	if _, ok := r.members[teamID][userID]; !ok {
		return ErrMemberNotFound
	}
	delete(r.members[teamID], userID)
	return nil
}

func (r *fakeTeamRepository) CountOwners(_ context.Context, teamID uint) (int64, error) {
	var count int64
	for _, member := range r.members[teamID] {
		if member.Role == RoleOwner {
			count++
		}
	}
	return count, nil
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

func TestOwnerInvitesRegisteredUser(t *testing.T) {
	repo := newFakeTeamRepository()
	repo.usersByEmail["member@example.com"] = UserBrief{ID: 8, Email: "member@example.com", Name: "Member"}
	svc := NewService(repo)
	created, err := svc.CreateTeam(context.Background(), 7, CreateTeamInput{Name: "Core"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}

	invitation, err := svc.InviteMember(context.Background(), 7, created.ID, InviteMemberInput{
		Email: " member@example.com ",
		Role:  RoleMember,
	})
	if err != nil {
		t.Fatalf("invite member: %v", err)
	}

	if invitation.ID == 0 || invitation.Email != "member@example.com" || invitation.Role != RoleMember || invitation.Status != InvitationPending {
		t.Fatalf("unexpected invitation: %#v", invitation)
	}
}

func TestInviteMemberRejectsNonOwner(t *testing.T) {
	repo := newFakeTeamRepository()
	repo.usersByEmail["member@example.com"] = UserBrief{ID: 8, Email: "member@example.com", Name: "Member"}
	svc := NewService(repo)
	created, err := svc.CreateTeam(context.Background(), 7, CreateTeamInput{Name: "Core"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	repo.members[created.ID][8] = &TeamMember{TeamID: created.ID, UserID: 8, Role: RoleMember}

	_, err = svc.InviteMember(context.Background(), 8, created.ID, InviteMemberInput{
		Email: "new@example.com",
		Role:  RoleMember,
	})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestAcceptInvitationAddsCurrentUserAsMember(t *testing.T) {
	repo := newFakeTeamRepository()
	repo.usersByEmail["member@example.com"] = UserBrief{ID: 8, Email: "member@example.com", Name: "Member"}
	svc := NewService(repo)
	created, err := svc.CreateTeam(context.Background(), 7, CreateTeamInput{Name: "Core"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	invitation, err := svc.InviteMember(context.Background(), 7, created.ID, InviteMemberInput{Email: "member@example.com", Role: RoleMember})
	if err != nil {
		t.Fatalf("invite member: %v", err)
	}

	accepted, err := svc.AcceptInvitation(context.Background(), 8, "member@example.com", invitation.ID)
	if err != nil {
		t.Fatalf("accept invitation: %v", err)
	}

	if accepted.Status != InvitationAccepted || repo.members[created.ID][8].Role != RoleMember {
		t.Fatalf("expected accepted member invitation, got %#v member=%#v", accepted, repo.members[created.ID][8])
	}
}

func TestCannotDemoteLastOwner(t *testing.T) {
	repo := newFakeTeamRepository()
	svc := NewService(repo)
	created, err := svc.CreateTeam(context.Background(), 7, CreateTeamInput{Name: "Core"})
	if err != nil {
		t.Fatalf("create team: %v", err)
	}

	_, err = svc.UpdateMemberRole(context.Background(), 7, created.ID, 7, UpdateMemberRoleInput{Role: RoleMember})
	if !errors.Is(err, ErrLastOwner) {
		t.Fatalf("expected ErrLastOwner, got %v", err)
	}
}
