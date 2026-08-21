import { create } from 'zustand';
import { createJSONStorage, persist } from 'zustand/middleware';

import { secureStorage } from '@/lib/secure-storage';
import type { AuthUser } from '@/types/auth';

interface AuthState {
  user: AuthUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  // False until the persisted session has been read from SecureStore, so
  // the root layout can hold navigation instead of flashing the login
  // screen for an already-authenticated user.
  hasHydrated: boolean;
  setSession: (user: AuthUser, accessToken: string, refreshToken: string) => void;
  setAccessToken: (accessToken: string) => void;
  clearSession: () => void;
  setHasHydrated: (value: boolean) => void;
}

// zustand's persist middleware runs hydration synchronously as part of the
// create() call below, before the `useAuthStore` export binding exists yet
// — referencing `useAuthStore` from inside onRehydrateStorage (a sibling of
// the state creator, not nested in it) throws a temporal-dead-zone
// ReferenceError. Capturing `set` here instead sidesteps that entirely.
let setAuthState: ((partial: Partial<AuthState>) => void) | undefined;

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => {
      setAuthState = set;
      return {
        user: null,
        accessToken: null,
        refreshToken: null,
        hasHydrated: false,
        setSession: (user, accessToken, refreshToken) => set({ user, accessToken, refreshToken }),
        setAccessToken: (accessToken) => set({ accessToken }),
        clearSession: () => set({ user: null, accessToken: null, refreshToken: null }),
        setHasHydrated: (value) => set({ hasHydrated: value }),
      };
    },
    {
      name: 'fieldsync-auth',
      storage: createJSONStorage(() => secureStorage),
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
      }),
      // zustand calls this with `state: undefined` whenever rehydration
      // throws for any reason (corrupted stored JSON, a storage adapter
      // bug, etc.) — relying only on `state?.setHasHydrated(true)` would
      // then silently never run, freezing the whole app on the loading
      // screen forever with no visible error. Falling back to the captured
      // `set` means a failed rehydration still unblocks the app — worst
      // case you're logged out, not stuck.
      onRehydrateStorage: () => (state, error) => {
        if (error) {
          console.warn('Failed to rehydrate persisted auth session', error);
        }
        if (state) {
          state.setHasHydrated(true);
        } else {
          setAuthState?.({ hasHydrated: true });
        }
      },
    },
  ),
);
