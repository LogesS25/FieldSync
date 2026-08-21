import { apiRequest } from '@/lib/api-client';
import type { AttendanceRecord, AttendanceSession } from '@/types/fieldwork';

export function listAttendance(): Promise<AttendanceRecord[]> {
  return apiRequest<AttendanceRecord[]>('/attendance');
}

export function listPendingAttendance(): Promise<AttendanceRecord[]> {
  return apiRequest<AttendanceRecord[]>('/attendance/pending');
}

export function createAttendance(input: {
  attendanceDate: string;
  session: AttendanceSession;
  hours?: number;
}): Promise<AttendanceRecord> {
  return apiRequest<AttendanceRecord>('/attendance', { method: 'POST', body: input });
}

export function agencyReviewAttendance(id: string, decision: 'approved' | 'rejected'): Promise<AttendanceRecord> {
  return apiRequest<AttendanceRecord>(`/attendance/${id}/agency-review`, { method: 'POST', body: { decision } });
}

export function facultyReviewAttendance(id: string, decision: 'approved' | 'rejected'): Promise<AttendanceRecord> {
  return apiRequest<AttendanceRecord>(`/attendance/${id}/faculty-review`, { method: 'POST', body: { decision } });
}
