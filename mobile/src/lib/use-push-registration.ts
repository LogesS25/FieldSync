import { useEffect } from 'react';

import { registerForPushNotificationsAsync } from '@/lib/push-notifications';
import * as pushTokensService from '@/services/push-tokens';
import { useAuthStore } from '@/stores/auth-store';

// Registers this device's Expo push token with the backend once the user is
// signed in. Runs once per sign-in (not on every render) and is entirely
// best-effort — see push-notifications.ts for why a missing token is
// expected and not an error.
export function usePushRegistration() {
  const userId = useAuthStore((state) => state.user?.id);
  const setPushToken = useAuthStore((state) => state.setPushToken);

  useEffect(() => {
    if (!userId) return;

    let cancelled = false;
    (async () => {
      const token = await registerForPushNotificationsAsync();
      if (!token || cancelled) return;
      try {
        await pushTokensService.registerPushToken(token);
        if (!cancelled) setPushToken(token);
      } catch {
        // Best-effort — a failed registration call shouldn't surface to the user.
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [userId, setPushToken]);
}
