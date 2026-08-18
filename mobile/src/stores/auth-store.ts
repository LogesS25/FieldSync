import { create } from 'zustand';

import type { AuthUser } from '@/types/auth';

// Placeholder shape only — real login, token refresh, and SecureStore
// persistence are Phase 2 (Authentication & User Foundation) work.
interface AuthState {
  user: AuthUser | null;
  accessToken: string | null;
  setSession: (user: AuthUser, accessToken: string) => void;
  clearSession: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  accessToken: null,
  setSession: (user, accessToken) => set({ user, accessToken }),
  clearSession: () => set({ user: null, accessToken: null }),
}));
