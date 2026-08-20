import { apiRequest } from '@/lib/api-client';
import type { AttendanceRecord, AttendanceSummary } from '@/types/fieldwork';

export function listAttendance(): Promise<AttendanceRecord[]> {
  return apiRequest<AttendanceRecord[]>('/attendance');
}

export function getAttendanceSummary(): Promise<AttendanceSummary> {
  return apiRequest<AttendanceSummary>('/attendance/summary');
}

export function createAttendance(input: { attendanceDate: string; hours: number }): Promise<AttendanceRecord> {
  return apiRequest<AttendanceRecord>('/attendance', { method: 'POST', body: input });
}
