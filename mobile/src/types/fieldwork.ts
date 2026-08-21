export type VerificationStatus = 'pending' | 'verified' | 'rejected';
export type ReviewDecision = 'pending' | 'approved' | 'rejected';
export type AttendanceSession = 'morning' | 'evening';
export type TeamRequestDecision = 'pending' | 'accepted' | 'rejected';

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
  session: AttendanceSession;
  hours: number | null;
  agencyStatus: ReviewDecision;
  facultyStatus: ReviewDecision;
}

export interface ConsolidatedReport {
  id: string;
  practicumId: string;
  summary: string;
  agencyStatus: ReviewDecision;
  facultyStatus: ReviewDecision;
  submittedAt: string;
}

export interface TeamRequest {
  id: string;
  studentId: string;
  agencyId: string;
  facultySupervisorId: string;
  agencySupervisorId: string;
  fieldworkDescription: string;
  startDate: string;
  facultyDecision: TeamRequestDecision;
  agencyDecision: TeamRequestDecision;
  formedPracticumId: string | null;
}
