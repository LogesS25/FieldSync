import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Text, View } from 'react-native';

import { Card } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import { LoadingState } from '@/components/ui/loading-state';
import { PageHeader } from '@/components/ui/page-header';
import { PrimaryButton } from '@/components/ui/primary-button';
import { ScreenContainer } from '@/components/ui/screen-container';
import * as attendanceService from '@/services/attendance';
import { useAuthStore } from '@/stores/auth-store';

export default function SupervisorAttendance() {
  const queryClient = useQueryClient();
  const user = useAuthStore((state) => state.user);
  const isAgency = user?.role === 'agency_supervisor';

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['attendance', 'pending'],
    queryFn: attendanceService.listPendingAttendance,
  });

  const mutation = useMutation({
    mutationFn: ({ id, decision }: { id: string; decision: 'approved' | 'rejected' }) =>
      isAgency ? attendanceService.agencyReviewAttendance(id, decision) : attendanceService.facultyReviewAttendance(id, decision),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['attendance', 'pending'] }),
  });

  return (
    <ScreenContainer scroll>
      <PageHeader
        icon="checkmark-done-circle-outline"
        title="Attendance Review"
        description={
          isAgency ? 'Review attendance before it goes to the faculty supervisor.' : 'Attendance already approved by the agency supervisor.'
        }
      />

      {isLoading ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState onRetry={() => refetch()} />
      ) : !data || data.length === 0 ? (
        <EmptyState
          icon="checkmark-done-outline"
          title="Nothing to review"
          description="Attendance awaiting your approval will show up here."
        />
      ) : (
        <View className="gap-3">
          {data.map((record) => (
            <Card key={record.id}>
              <Text className="mb-3 font-medium text-slate-800">
                {record.attendanceDate} · <Text className="capitalize">{record.session}</Text>
                {record.hours ? ` · ${record.hours}h` : ''}
              </Text>
              <View className="flex-row gap-3">
                <View className="flex-1">
                  <PrimaryButton
                    label="Approve"
                    onPress={() => mutation.mutate({ id: record.id, decision: 'approved' })}
                    disabled={mutation.isPending}
                  />
                </View>
                <View className="flex-1">
                  <PrimaryButton
                    label="Reject"
                    variant="danger"
                    onPress={() => mutation.mutate({ id: record.id, decision: 'rejected' })}
                    disabled={mutation.isPending}
                  />
                </View>
              </View>
            </Card>
          ))}
        </View>
      )}
    </ScreenContainer>
  );
}
