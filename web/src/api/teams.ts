import { api } from './http';
import type {
  CreateTeamPayload,
  InviteTeamMemberPayload,
  Team,
  TeamInvitation,
  TeamMember,
  UpdateTeamMemberRolePayload
} from './types';

export const teamsApi = {
  create(payload: CreateTeamPayload) {
    return api.post<Team>('/teams', payload);
  },
  list() {
    return api.get<Team[]>('/teams');
  },
  get(id: number) {
    return api.get<Team>(`/teams/${id}`);
  },
  listMembers(teamID: number) {
    return api.get<TeamMember[]>(`/teams/${teamID}/members`);
  },
  inviteMember(teamID: number, payload: InviteTeamMemberPayload) {
    return api.post<TeamInvitation>(`/teams/${teamID}/invitations`, payload);
  },
  listInvitations() {
    return api.get<TeamInvitation[]>('/teams/invitations');
  },
  acceptInvitation(invitationID: number) {
    return api.post<TeamInvitation>(`/teams/invitations/${invitationID}/accept`);
  },
  updateMemberRole(teamID: number, userID: number, payload: UpdateTeamMemberRolePayload) {
    return api.patch<TeamMember>(`/teams/${teamID}/members/${userID}/role`, payload);
  },
  removeMember(teamID: number, userID: number) {
    return api.delete<{ removed: boolean }>(`/teams/${teamID}/members/${userID}`);
  }
};
