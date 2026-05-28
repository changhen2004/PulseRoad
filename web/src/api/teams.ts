import { api } from './http';
import type { CreateTeamPayload, Team } from './types';

export const teamsApi = {
  create(payload: CreateTeamPayload) {
    return api.post<Team>('/teams', payload);
  },
  list() {
    return api.get<Team[]>('/teams');
  },
  get(id: number) {
    return api.get<Team>(`/teams/${id}`);
  }
};
