import { Redirect } from 'expo-router';
import { ActivityIndicator, View } from 'react-native';

import { useAuthStore } from '@/stores/auth-store';

// Entry redirect: unauthenticated users go to login, authenticated users go
// to the dashboard for their role. Faculty and Agency Supervisors share the
// (supervisor) route group — their capability lists in the requirements are
// structurally similar, and per-screen actions are gated by role.
export default function Index() {
  const hasHydrated = useAuthStore((state) => state.hasHydrated);
  const user = useAuthStore((state) => state.user);

  if (!hasHydrated) {
    // Wait for the SecureStore-persisted session to load before deciding
    // where to send an already-authenticated user, so they don't flash
    // through the login screen on every cold start.
    return (
      <View className="flex-1 items-center justify-center bg-white">
        <ActivityIndicator />
      </View>
    );
  }

  if (!user) {
    return <Redirect href="/(auth)/login" />;
  }

  if (user.role === 'student') {
    return <Redirect href="/(student)/dashboard" />;
  }

  return <Redirect href="/(supervisor)/dashboard" />;
}
