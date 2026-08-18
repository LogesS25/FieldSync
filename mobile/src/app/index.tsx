import { Redirect } from 'expo-router';
import { ActivityIndicator, Text, View } from 'react-native';

import { LogoutButton } from '@/components/logout-button';
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

  if (user.role === 'faculty_supervisor' || user.role === 'agency_supervisor') {
    return <Redirect href="/(supervisor)/dashboard" />;
  }

  // Administrator accounts exist (created directly in Postgres — see
  // docs/ARCHITECTURE.md §5b) but the mobile app doesn't have a UI for
  // that role (Phase 9 builds a separate web dashboard). Without this
  // branch, an admin would bounce forever between "/" and the (supervisor)
  // group's RequireRole redirect back to "/".
  return (
    <View className="flex-1 items-center justify-center bg-white px-6">
      <Text className="text-center text-base font-medium text-slate-700">
        Administrator accounts aren&apos;t supported in the mobile app yet.
      </Text>
      <Text className="mt-2 text-center text-sm text-slate-500">
        Use the web dashboard once it&apos;s available.
      </Text>
      <View className="mt-6">
        <LogoutButton />
      </View>
    </View>
  );
}
