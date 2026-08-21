import { apiRequest } from '@/lib/api-client';
import type { Agency, FieldworkComponent, Institution } from '@/types/directory';

// Unauthenticated — used by the registration screen before the user has an
// account (§/public/institutions, /public/agencies on the backend).
export function listPublicInstitutions(): Promise<Institution[]> {
  return apiRequest<Institution[]>('/public/institutions');
}

export function listPublicAgencies(): Promise<Agency[]> {
  return apiRequest<Agency[]>('/public/agencies');
}

// Authenticated — a student/faculty supervisor browsing agencies within
// their own university (e.g. when forming a practicum team request).
export function listMyAgencies(): Promise<Agency[]> {
  return apiRequest<Agency[]>('/agencies/mine');
}

// University-defined list a student picks from when forming a practicum
// team request (business requirements §7).
export function listMyFieldworkComponents(): Promise<FieldworkComponent[]> {
  return apiRequest<FieldworkComponent[]>('/fieldwork-components/mine');
}
