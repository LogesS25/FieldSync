import { apiRequest } from '@/lib/api-client';
import type { AuthUser } from '@/types/auth';
import type { TeamRequest } from '@/types/fieldwork';

export function createTeamRequest(input: {
  agencyId: string;
  facultySupervisorId: string;
  agencySupervisorId: string;
  fieldworkDescription: string;
  startDate: string;
}): Promise<TeamRequest> {
  return apiRequest<TeamRequest>('/team-requests', { method: 'POST', body: input });
}

export function listMyTeamRequests(): Promise<TeamRequest[]> {
  return apiRequest<TeamRequest[]>('/team-requests/me');
}

export function listPendingTeamRequests(): Promise<TeamRequest[]> {
  return apiRequest<TeamRequest[]>('/team-requests/pending');
}

export function respondToTeamRequest(id: string, decision: 'accepted' | 'rejected'): Promise<TeamRequest> {
  return apiRequest<TeamRequest>(`/team-requests/${id}/respond`, { method: 'POST', body: { decision } });
}

export function listMyFacultySupervisors(): Promise<AuthUser[]> {
  return apiRequest<AuthUser[]>('/faculty-supervisors/mine');
}

export function listAgencySupervisors(agencyId: string): Promise<AuthUser[]> {
  return apiRequest<AuthUser[]>(`/agency-supervisors?agencyId=${agencyId}`);
}
