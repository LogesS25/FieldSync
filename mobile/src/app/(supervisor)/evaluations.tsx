import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Text, View } from 'react-native';

import { Card } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import { LoadingState } from '@/components/ui/loading-state';
import { PageHeader } from '@/components/ui/page-header';
import { PrimaryButton } from '@/components/ui/primary-button';
import { ScreenContainer } from '@/components/ui/screen-container';
import * as reportsService from '@/services/reports';
import { useAuthStore } from '@/stores/auth-store';

export default function SupervisorConsolidatedReports() {
  const queryClient = useQueryClient();
  const user = useAuthStore((state) => state.user);
  const isAgency = user?.role === 'agency_supervisor';

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['consolidated-reports', 'pending'],
    queryFn: reportsService.listPendingConsolidatedReports,
  });

  const mutation = useMutation({
    mutationFn: ({ id, decision }: { id: string; decision: 'approved' | 'rejected' }) =>
      isAgency ? reportsService.agencyReviewReport(id, decision) : reportsService.facultyReviewReport(id, decision),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['consolidated-reports', 'pending'] }),
  });

  return (
    <ScreenContainer scroll>
      <PageHeader
        icon="bar-chart-outline"
        title="Consolidated Reports"
        description={isAgency ? 'Review before the faculty supervisor does.' : 'Already approved by the agency supervisor.'}
      />

      {isLoading ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState onRetry={() => refetch()} />
      ) : !data || data.length === 0 ? (
        <EmptyState
          icon="checkmark-done-outline"
          title="Nothing to review"
          description="Reports awaiting your approval will show up here."
        />
      ) : (
        <View className="gap-3">
          {data.map((report) => (
            <Card key={report.id}>
              <Text className="mb-3 text-sm text-slate-700">{report.summary}</Text>
              <View className="flex-row gap-3">
                <View className="flex-1">
                  <PrimaryButton
                    label="Approve"
                    onPress={() => mutation.mutate({ id: report.id, decision: 'approved' })}
                    disabled={mutation.isPending}
                  />
                </View>
                <View className="flex-1">
                  <PrimaryButton
                    label="Reject"
                    variant="danger"
                    onPress={() => mutation.mutate({ id: report.id, decision: 'rejected' })}
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
