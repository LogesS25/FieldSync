export type UserRole = 'student' | 'faculty_supervisor' | 'agency_supervisor';

export interface AuthUser {
  id: string;
  email: string;
  fullName: string;
  role: UserRole;
  institutionId: string | null;
  agencyId: string | null;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
}
