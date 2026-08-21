import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Pressable, Text, View } from 'react-native';

import { Card } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { PrimaryButton } from '@/components/ui/primary-button';
import { ScreenContainer } from '@/components/ui/screen-container';
import { openAuthenticatedFile } from '@/lib/open-file';
import * as dailyReportsService from '@/services/dailyreports';
import { useAuthStore } from '@/stores/auth-store';
import type { DailyReport } from '@/types/dailyreport';

export default function SupervisorDailyReports() {
  const queryClient = useQueryClient();
  const accessToken = useAuthStore((state) => state.accessToken);
  const user = useAuthStore((state) => state.user);
  const isAgency = user?.role === 'agency_supervisor';

  const { data, isLoading } = useQuery({
    queryKey: ['daily-reports', 'pending'],
    queryFn: dailyReportsService.listPendingDailyReports,
  });

  const mutation = useMutation({
    mutationFn: ({ id, decision }: { id: string; decision: 'approved' | 'rejected' }) =>
      isAgency
        ? dailyReportsService.agencyReviewDailyReport(id, decision)
        : dailyReportsService.facultyReviewDailyReport(id, decision),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['daily-reports', 'pending'] }),
  });

  const viewFile = async (report: DailyReport) => {
    if (!accessToken) return;
    try {
      await openAuthenticatedFile(dailyReportsService.dailyReportFileUrl(report.id), accessToken, report.filename);
    } catch {
      // Best-effort viewer — a failed open isn't worth a blocking error UI.
    }
  };

  return (
    <ScreenContainer scroll>
      <Text className="mb-1 text-2xl font-bold text-slate-900">Daily Reports</Text>
      <Text className="mb-6 text-sm text-slate-500">
        {isAgency ? 'Review before the faculty supervisor does.' : 'Already approved by the agency supervisor.'}
      </Text>

      {isLoading ? null : !data || data.length === 0 ? (
        <EmptyState title="Nothing to review" description="Reports awaiting your approval will show up here." />
      ) : (
        <View className="gap-3">
          {data.map((report) => (
            <Card key={report.id}>
              <Text className="mb-1 font-medium text-slate-800">{report.reportDate}</Text>
              <Pressable onPress={() => viewFile(report)} className="mb-3">
                <Text className="text-sm text-brand-600">{report.filename}</Text>
              </Pressable>
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
