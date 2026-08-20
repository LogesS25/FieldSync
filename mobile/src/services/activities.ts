import { apiRequest } from '@/lib/api-client';
import type { FieldActivity } from '@/types/fieldwork';

export function listFieldActivities(): Promise<FieldActivity[]> {
  return apiRequest<FieldActivity[]>('/field-activities');
}

export function createFieldActivity(input: { activityDate: string; description: string }): Promise<FieldActivity> {
  return apiRequest<FieldActivity>('/field-activities', { method: 'POST', body: input });
}
