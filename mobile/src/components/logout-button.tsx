import { Ionicons } from '@expo/vector-icons';
import { Pressable, Text } from 'react-native';

import * as authService from '@/services/auth';
import * as pushTokensService from '@/services/push-tokens';
import { useAuthStore } from '@/stores/auth-store';

interface LogoutButtonProps {
  fullWidth?: boolean;
}

export function LogoutButton({ fullWidth }: LogoutButtonProps = {}) {
  const refreshToken = useAuthStore((state) => state.refreshToken);
  const pushToken = useAuthStore((state) => state.pushToken);
  const clearSession = useAuthStore((state) => state.clearSession);

  const handleLogout = async () => {
    if (pushToken) {
      // Best-effort: stop this device from receiving push for the
      // account that's signing out, but never block logout on it.
      await pushTokensService.unregisterPushToken(pushToken).catch(() => undefined);
    }
    if (refreshToken) {
      // Best-effort: revoke the refresh token server-side, but always clear
      // the local session even if the request fails (e.g. offline).
      await authService.logout(refreshToken).catch(() => undefined);
    }
    clearSession();
  };

  return (
    <Pressable
      className={`flex-row items-center justify-center gap-2 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 active:opacity-70 ${fullWidth ? 'w-full' : 'self-start'}`}
      onPress={handleLogout}
    >
      <Ionicons name="log-out-outline" size={17} color="#e11d48" />
      <Text className="text-sm font-semibold text-rose-600">Sign out</Text>
    </Pressable>
  );
}
