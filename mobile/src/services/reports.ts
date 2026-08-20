import { apiRequest } from '@/lib/api-client';
import type { WeeklyReport } from '@/types/fieldwork';

export function listWeeklyReports(): Promise<WeeklyReport[]> {
  return apiRequest<WeeklyReport[]>('/weekly-reports');
}

export function submitWeeklyReport(input: {
  weekStartDate: string;
  weekEndDate: string;
  summary: string;
}): Promise<WeeklyReport> {
  return apiRequest<WeeklyReport>('/weekly-reports', { method: 'POST', body: input });
}
