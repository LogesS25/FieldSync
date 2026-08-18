import { apiRequest } from '@/lib/api-client';
import type { AssignedStudent, PracticumSummary } from '@/types/practicum';

export function getMyPracticum(): Promise<PracticumSummary> {
  return apiRequest<PracticumSummary>('/practicums/me');
}

export function listMyStudents(): Promise<AssignedStudent[]> {
  return apiRequest<AssignedStudent[]>('/students');
}
