import { Pressable, Text } from 'react-native';

import * as authService from '@/services/auth';
import { useAuthStore } from '@/stores/auth-store';

export function LogoutButton() {
  const refreshToken = useAuthStore((state) => state.refreshToken);
  const clearSession = useAuthStore((state) => state.clearSession);

  const handleLogout = async () => {
    if (refreshToken) {
      // Best-effort: revoke the refresh token server-side, but always clear
      // the local session even if the request fails (e.g. offline).
      await authService.logout(refreshToken).catch(() => undefined);
    }
    clearSession();
  };

  return (
    <Pressable className="mt-6 rounded-lg border border-slate-300 px-4 py-2" onPress={handleLogout}>
      <Text className="text-slate-700">Sign out</Text>
    </Pressable>
  );
}
