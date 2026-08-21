import { useQuery } from '@tanstack/react-query';
import { ActivityIndicator, Text, View } from 'react-native';

import { Avatar } from '@/components/ui/avatar';
import { Badge, type BadgeTone } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { ScreenContainer } from '@/components/ui/screen-container';
import { LogoutButton } from '@/components/logout-button';
import { ApiError } from '@/lib/api-client';
import * as practicumService from '@/services/practicums';
import { useAuthStore } from '@/stores/auth-store';
import type { PracticumStatus } from '@/types/practicum';

const STATUS_TONE: Record<PracticumStatus, BadgeTone> = {
  active: 'success',
  completed: 'neutral',
  terminated: 'danger',
};

const STATUS_LABEL: Record<PracticumStatus, string> = {
  active: 'Active',
  completed: 'Completed',
  terminated: 'Terminated',
};

export default function StudentDashboard() {
  const user = useAuthStore((state) => state.user);

  const { data, isLoading, error } = useQuery({
    queryKey: ['practicums', 'me'],
    queryFn: practicumService.getMyPracticum,
    retry: (failureCount, err) => !(err instanceof ApiError && err.status === 404) && failureCount < 1,
  });

  const hasNoPracticum = error instanceof ApiError && error.status === 404;

  return (
    <ScreenContainer scroll>
      <Text className="text-sm font-medium text-brand-600">Welcome back</Text>
      <Text className="mb-6 text-2xl font-bold text-slate-900">{user?.fullName ?? 'Student'}</Text>

      {isLoading ? (
        <View className="items-center py-10">
          <ActivityIndicator color="#4f46e5" />
        </View>
      ) : hasNoPracticum ? (
        <EmptyState
          title="No active practicum yet"
          description="Send a team request from the Team tab once you've picked an agency and supervisors."
        />
      ) : data ? (
        <Card>
          <View className="mb-4 flex-row items-center justify-between">
            <Text className="text-lg font-semibold text-slate-900">Your Practicum</Text>
            <Badge label={STATUS_LABEL[data.status]} tone={STATUS_TONE[data.status]} />
          </View>

          <View className="gap-3">
            <InfoRow label="Institution" value={data.institutionName} />
            <InfoRow label="Agency" value={data.agencyName ?? 'Not yet placed'} />
            <InfoRow label="Started" value={data.startDate} />
            {data.endDate ? <InfoRow label="Ends" value={data.endDate} /> : null}
          </View>

          {data.supervisors.length > 0 ? (
            <View className="mt-5 border-t border-slate-100 pt-4">
              <Text className="mb-3 text-sm font-medium text-slate-500">Supervisors</Text>
              <View className="gap-3">
                {data.supervisors.map((supervisor) => (
                  <View key={supervisor.id} className="flex-row items-center gap-3">
                    <Avatar name={supervisor.fullName} size="sm" />
                    <View>
                      <Text className="font-medium text-slate-800">{supervisor.fullName}</Text>
                      <Text className="text-xs capitalize text-slate-500">
                        {supervisor.role.replace('_', ' ')}
                      </Text>
                    </View>
                  </View>
                ))}
              </View>
            </View>
          ) : null}
        </Card>
      ) : null}

      <View className="mt-8 items-start">
        <LogoutButton />
      </View>
    </ScreenContainer>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <View className="flex-row items-center justify-between">
      <Text className="text-sm text-slate-500">{label}</Text>
      <Text className="text-sm font-medium text-slate-800">{value}</Text>
    </View>
  );
}
