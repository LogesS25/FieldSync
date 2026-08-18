import '@/global.css';

import { Stack } from 'expo-router';

import { AppQueryProvider } from '@/lib/query-client';

export default function RootLayout() {
  return (
    <AppQueryProvider>
      <Stack screenOptions={{ headerShown: false }}>
        <Stack.Screen name="(auth)" />
        <Stack.Screen name="(student)" />
        <Stack.Screen name="(supervisor)" />
      </Stack>
    </AppQueryProvider>
  );
}
