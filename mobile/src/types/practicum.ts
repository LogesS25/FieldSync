export type PracticumStatus = 'active' | 'completed' | 'terminated';

export interface PracticumSupervisor {
  id: string;
  fullName: string;
  role: string;
}

export interface PracticumSummary {
  practicumId: string;
  status: PracticumStatus;
  startDate: string;
  endDate: string | null;
  institutionName: string;
  agencyId: string | null;
  agencyName: string | null;
  placementStartDate: string | null;
  placementEndDate: string | null;
  supervisors: PracticumSupervisor[];
}

export interface AssignedStudent {
  studentId: string;
  studentName: string;
  studentEmail: string;
  practicumId: string;
  practicumStatus: PracticumStatus;
  startDate: string;
  endDate: string | null;
  institutionName: string;
  agencyId: string | null;
  agencyName: string | null;
}
