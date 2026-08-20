export type VerificationStatus = 'pending' | 'verified' | 'rejected';
export type ReportStatus = 'submitted' | 'reviewed';

export interface FieldActivity {
  id: string;
  practicumId: string;
  activityDate: string;
  description: string;
  verificationStatus: VerificationStatus;
}

export interface AttendanceRecord {
  id: string;
  practicumId: string;
  attendanceDate: string;
  hours: number;
  verificationStatus: VerificationStatus;
}

export interface AttendanceSummary {
  totalHours: number;
}

export interface WeeklyReport {
  id: string;
  practicumId: string;
  weekStartDate: string;
  weekEndDate: string;
  summary: string;
  status: ReportStatus;
}
