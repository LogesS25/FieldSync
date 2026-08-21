import { API_URL } from '@/lib/env';
import { apiRequest, apiUpload } from '@/lib/api-client';
import type { DailyReport } from '@/types/dailyreport';

export interface PickedFile {
  uri: string;
  name: string;
  mimeType: string;
}

export function submitDailyReport(input: { reportDate: string; file: PickedFile }): Promise<DailyReport> {
  const formData = new FormData();
  formData.append('reportDate', input.reportDate);
  formData.append('file', {
    uri: input.file.uri,
    name: input.file.name,
    type: input.file.mimeType,
  } as unknown as Blob);

  return apiUpload<DailyReport>('/daily-reports', formData);
}

export function listMyDailyReports(): Promise<DailyReport[]> {
  return apiRequest<DailyReport[]>('/daily-reports');
}

export function listPendingDailyReports(): Promise<DailyReport[]> {
  return apiRequest<DailyReport[]>('/daily-reports/pending');
}

export function agencyReviewDailyReport(id: string, decision: 'approved' | 'rejected'): Promise<DailyReport> {
  return apiRequest<DailyReport>(`/daily-reports/${id}/agency-review`, { method: 'POST', body: { decision } });
}

export function facultyReviewDailyReport(id: string, decision: 'approved' | 'rejected'): Promise<DailyReport> {
  return apiRequest<DailyReport>(`/daily-reports/${id}/faculty-review`, { method: 'POST', body: { decision } });
}

// The file endpoint requires an Authorization header, so it can't be a
// plain <a href>/Linking.openURL URL — callers fetch it with the current
// access token and hand the viewer a blob/data URL instead.
export function dailyReportFileUrl(id: string): string {
  return `${API_URL}/daily-reports/${id}/file`;
}
