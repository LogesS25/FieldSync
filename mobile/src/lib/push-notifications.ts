import Constants from 'expo-constants';
import * as Device from 'expo-device';
import * as Notifications from 'expo-notifications';
import { Platform } from 'react-native';

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowBanner: true,
    shouldShowList: true,
    shouldPlaySound: false,
    shouldSetBadge: false,
  }),
});

// Returns an Expo push token for this device, or null if push isn't
// available — which is the common case until this project is linked to an
// EAS project (`eas login` + `eas init`, done once by a human with an Expo
// account; see AGENTS.md) and/or when running in a plain simulator/Expo Go
// environment that doesn't support remote push. This is best-effort by
// design: a missing push token should never block login or break the app,
// it just means this device won't receive push notifications (in-app
// notifications still work regardless).
export async function registerForPushNotificationsAsync(): Promise<string | null> {
  if (!Device.isDevice) return null;

  try {
    const { status: existingStatus } = await Notifications.getPermissionsAsync();
    let finalStatus = existingStatus;
    if (existingStatus !== 'granted') {
      const { status } = await Notifications.requestPermissionsAsync();
      finalStatus = status;
    }
    if (finalStatus !== 'granted') return null;

    if (Platform.OS === 'android') {
      await Notifications.setNotificationChannelAsync('default', {
        name: 'default',
        importance: Notifications.AndroidImportance.DEFAULT,
      });
    }

    const projectId = Constants.expoConfig?.extra?.eas?.projectId as string | undefined;
    if (!projectId) return null;

    const { data } = await Notifications.getExpoPushTokenAsync({ projectId });
    return data;
  } catch {
    // Expo Go on Android (SDK 53+) and various dev-client edge cases throw
    // here rather than returning a clean unavailable state — best-effort
    // means swallowing this, not surfacing it as an app error.
    return null;
  }
}
