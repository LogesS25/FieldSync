import { useQuery } from '@tanstack/react-query';
import { Text, View } from 'react-native';

import { getHealth } from '@/services/health';

// Placeholder screen. Real login (form, validation, JWT exchange) is built
// in Phase 2 alongside the backend auth endpoints. The health check below
// exists only to prove the mobile app can reach the Go API during Phase 1
// setup verification.
export default function LoginScreen() {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['health'],
    queryFn: getHealth,
  });

  return (
    <View className="flex-1 items-center justify-center bg-white px-6">
      <Text className="text-2xl font-semibold text-slate-900">FieldSync</Text>
      <Text className="mt-2 text-center text-slate-500">
        Login will be implemented in Phase 2 (Authentication & User Foundation).
      </Text>
      <Text className="mt-6 text-sm text-slate-400">
        API status:{' '}
        {isLoading ? 'checking…' : isError ? 'unreachable' : `${data?.status} (db: ${data?.database})`}
      </Text>
    </View>
  );
}
