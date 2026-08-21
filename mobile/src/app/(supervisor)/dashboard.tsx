import { useQuery } from '@tanstack/react-query';
import { router } from 'expo-router';
import { Text, View } from 'react-native';

import { Avatar } from '@/components/ui/avatar';
import { Card } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import { LoadingState } from '@/components/ui/loading-state';
import { PageHeader } from '@/components/ui/page-header';
import { ScreenContainer } from '@/components/ui/screen-container';
import { StatCard } from '@/components/ui/stat-card';
import * as attendanceService from '@/services/attendance';
import * as dailyReportsService from '@/services/dailyreports';
import * as practicumService from '@/services/practicums';
import * as reportsService from '@/services/reports';
import { useAuthStore } from '@/stores/auth-store';

export default function SupervisorDashboard() {
  const user = useAuthStore((state) => state.user);

  const students = useQuery({ queryKey: ['students'], queryFn: practicumService.listMyStudents });
  const pendingAttendance = useQuery({ queryKey: ['attendance', 'pending'], queryFn: attendanceService.listPendingAttendance });
  const pendingReports = useQuery({ queryKey: ['daily-reports', 'pending'], queryFn: dailyReportsService.listPendingDailyReports });
  const pendingConsolidated = useQuery({
    queryKey: ['consolidated-reports', 'pending'],
    queryFn: reportsService.listPendingConsolidatedReports,
  });

  const isLoading =
    students.isLoading || pendingAttendance.isLoading || pendingReports.isLoading || pendingConsolidated.isLoading;
  const hasError = students.isError || pendingAttendance.isError || pendingReports.isError || pendingConsolidated.isError;

  const retryAll = () => {
    students.refetch();
    pendingAttendance.refetch();
    pendingReports.refetch();
    pendingConsolidated.refetch();
  };

  return (
    <ScreenContainer scroll>
      <PageHeader
        icon="home-outline"
        title={`Welcome, ${user?.fullName?.split(' ')[0] ?? 'Supervisor'}`}
        description="Here's what needs your attention today."
      />

      {isLoading ? (
        <LoadingState />
      ) : hasError ? (
        <ErrorState onRetry={retryAll} />
      ) : (
        <>
          <View className="mb-6 flex-row flex-wrap gap-3">
            <StatCard
              label="Students"
              value={students.data?.length ?? 0}
              icon="people-outline"
              onPress={() => router.push('/(supervisor)/students')}
            />
            <StatCard
              label="Attendance to review"
              value={pendingAttendance.data?.length ?? 0}
              icon="checkmark-done-circle-outline"
              onPress={() => router.push('/(supervisor)/attendance')}
            />
            <StatCard
              label="Daily reports to review"
              value={pendingReports.data?.length ?? 0}
              icon="document-text-outline"
              onPress={() => router.push('/(supervisor)/activities')}
            />
            <StatCard
              label="Consolidated reports"
              value={pendingConsolidated.data?.length ?? 0}
              icon="bar-chart-outline"
              onPress={() => router.push('/(supervisor)/evaluations')}
            />
          </View>

          <Text className="mb-3 text-sm font-semibold text-slate-500">Your students</Text>
          {!students.data || students.data.length === 0 ? (
            <EmptyState
              icon="people-outline"
              title="No students assigned yet"
              description="Students will appear here once an administrator assigns you to their practicum."
            />
          ) : (
            <View className="gap-3">
              {students.data.slice(0, 4).map((student) => (
                <Card key={student.practicumId}>
                  <View className="flex-row items-center gap-3">
                    <Avatar name={student.studentName} size="sm" />
                    <View className="flex-1">
                      <Text className="font-medium text-slate-800">{student.studentName}</Text>
                      <Text className="text-xs text-slate-500">{student.agencyName ?? 'Not yet placed'}</Text>
                    </View>
                  </View>
                </Card>
              ))}
            </View>
          )}
        </>
      )}
    </ScreenContainer>
  );
}
