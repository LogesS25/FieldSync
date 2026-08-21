import { ApiError, apiRequest } from '@/lib/api-client';
import type { ConsolidatedReport } from '@/types/fieldwork';

export function submitConsolidatedReport(summary: string): Promise<ConsolidatedReport> {
  return apiRequest<ConsolidatedReport>('/consolidated-reports', { method: 'POST', body: { summary } });
}

export async function getMyConsolidatedReport(): Promise<ConsolidatedReport | null> {
  try {
    return await apiRequest<ConsolidatedReport>('/consolidated-reports/me');
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }
}

export function listPendingConsolidatedReports(): Promise<ConsolidatedReport[]> {
  return apiRequest<ConsolidatedReport[]>('/consolidated-reports/pending');
}

export function agencyReviewReport(id: string, decision: 'approved' | 'rejected'): Promise<ConsolidatedReport> {
  return apiRequest<ConsolidatedReport>(`/consolidated-reports/${id}/agency-review`, { method: 'POST', body: { decision } });
}

export function facultyReviewReport(id: string, decision: 'approved' | 'rejected'): Promise<ConsolidatedReport> {
  return apiRequest<ConsolidatedReport>(`/consolidated-reports/${id}/faculty-review`, { method: 'POST', body: { decision } });
}
