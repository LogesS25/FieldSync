import { Redirect } from 'expo-router';
import type { PropsWithChildren } from 'react';

import { useAuthStore } from '@/stores/auth-store';
import type { UserRole } from '@/types/auth';

interface RequireRoleProps extends PropsWithChildren {
  allowedRoles: UserRole[];
}

// Guards a route group: sends unauthenticated users to login, and signed-in
// users of the wrong role back to the entry redirect (which sends them to
// their own dashboard). Complements backend authorization — this is a UX
// guard, not the security boundary.
export function RequireRole({ allowedRoles, children }: RequireRoleProps) {
  const user = useAuthStore((state) => state.user);

  if (!user) {
    return <Redirect href="/(auth)/login" />;
  }

  if (!allowedRoles.includes(user.role)) {
    return <Redirect href="/" />;
  }

  return children;
}
