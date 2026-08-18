import { apiRequest } from '@/lib/api-client';
import type { AuthUser, UserRole } from '@/types/auth';

interface SessionResponse {
  accessToken: string;
  refreshToken: string;
  user: AuthUser;
}

export function login(email: string, password: string): Promise<SessionResponse> {
  return apiRequest<SessionResponse>('/auth/login', {
    method: 'POST',
    body: { email, password },
  });
}

export function register(input: {
  email: string;
  password: string;
  fullName: string;
  role: UserRole;
}): Promise<SessionResponse> {
  return apiRequest<SessionResponse>('/auth/register', {
    method: 'POST',
    body: input,
  });
}

export function logout(refreshToken: string): Promise<void> {
  return apiRequest<void>('/auth/logout', {
    method: 'POST',
    body: { refreshToken },
  });
}
