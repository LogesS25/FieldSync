import '@/global.css';

import { Stack } from 'expo-router';
import { GestureHandlerRootView } from 'react-native-gesture-handler';

import { AppQueryProvider } from '@/lib/query-client';

export default function RootLayout() {
  return (
    // The (student)/(supervisor) Drawer navigators need a GestureHandlerRootView
    // ancestor for their swipe-to-open gesture and internal panning — without
    // it the drawer still renders but gestures silently no-op.
    <GestureHandlerRootView style={{ flex: 1 }}>
      <AppQueryProvider>
        <Stack screenOptions={{ headerShown: false }}>
          <Stack.Screen name="(auth)" />
          <Stack.Screen name="(student)" />
          <Stack.Screen name="(supervisor)" />
        </Stack>
      </AppQueryProvider>
    </GestureHandlerRootView>
  );
}
