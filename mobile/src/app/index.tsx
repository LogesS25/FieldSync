import { Redirect } from 'expo-router';

import { useAuthStore } from '@/stores/auth-store';

// Entry redirect: unauthenticated users go to login, authenticated users go
// to the dashboard for their role. Faculty and Agency Supervisors share the
// (supervisor) route group — their capability lists in the requirements are
// structurally similar, and per-screen actions are gated by role.
export default function Index() {
  const user = useAuthStore((state) => state.user);

  if (!user) {
    return <Redirect href="/(auth)/login" />;
  }

  if (user.role === 'student') {
    return <Redirect href="/(student)/dashboard" />;
  }

  return <Redirect href="/(supervisor)/dashboard" />;
}
