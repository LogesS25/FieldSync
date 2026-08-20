import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Controller, useForm } from 'react-hook-form';
import { Text, View } from 'react-native';

import { DateInput } from '@/components/ui/date-input';
import { LabeledInput } from '@/components/ui/labeled-input';
import { PrimaryButton } from '@/components/ui/primary-button';
import { Badge, type BadgeTone } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { ScreenContainer } from '@/components/ui/screen-container';
import { weeklyReportSchema, type WeeklyReportFormValues } from '@/features/fieldwork/schemas';
import { ApiError } from '@/lib/api-client';
import * as reportsService from '@/services/reports';
import type { ReportStatus, WeeklyReport } from '@/types/fieldwork';

const STATUS_TONE: Record<ReportStatus, BadgeTone> = {
  submitted: 'info',
  reviewed: 'success',
};

export default function StudentReports() {
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['weekly-reports'],
    queryFn: reportsService.listWeeklyReports,
  });

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<WeeklyReportFormValues>({
    resolver: zodResolver(weeklyReportSchema),
    defaultValues: { weekStartDate: '', weekEndDate: '', summary: '' },
  });

  const mutation = useMutation({
    mutationFn: reportsService.submitWeeklyReport,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['weekly-reports'] });
      reset();
    },
  });

  const onSubmit = handleSubmit((values) => mutation.mutate(values));

  return (
    <ScreenContainer scroll>
      <Text className="mb-1 text-2xl font-bold text-slate-900">Weekly Reports</Text>
      <Text className="mb-6 text-sm text-slate-500">
        Submitted reports can&apos;t be edited afterward — double-check before submitting.
      </Text>

      <Card className="mb-6">
        <Controller
          control={control}
          name="weekStartDate"
          render={({ field: { onChange, value } }) => (
            <DateInput label="Week start" value={value} onChange={onChange} error={errors.weekStartDate?.message} />
          )}
        />
        <Controller
          control={control}
          name="weekEndDate"
          render={({ field: { onChange, value } }) => (
            <DateInput label="Week end" value={value} onChange={onChange} error={errors.weekEndDate?.message} />
          )}
        />
        <Controller
          control={control}
          name="summary"
          render={({ field: { onChange, onBlur, value } }) => (
            <LabeledInput
              label="Summary"
              placeholder="What did you accomplish this week?"
              multiline
              numberOfLines={4}
              onBlur={onBlur}
              onChangeText={onChange}
              value={value}
              error={errors.summary?.message}
            />
          )}
        />
        {mutation.isError ? (
          <Text className="mb-2 text-sm text-red-600">
            {mutation.error instanceof ApiError && mutation.error.status === 409
              ? 'A report for that week already exists, or you have no active practicum.'
              : 'Something went wrong. Please try again.'}
          </Text>
        ) : null}
        <PrimaryButton
          label={mutation.isPending ? 'Submitting…' : 'Submit Report'}
          onPress={onSubmit}
          disabled={mutation.isPending}
        />
      </Card>

      {isLoading ? null : !data || data.length === 0 ? (
        <EmptyState title="No reports submitted yet" description="Weekly reports you submit will show up here." />
      ) : (
        <View className="gap-3">
          {data.map((report) => (
            <ReportCard key={report.id} report={report} />
          ))}
        </View>
      )}
    </ScreenContainer>
  );
}

function ReportCard({ report }: { report: WeeklyReport }) {
  return (
    <Card>
      <View className="mb-2 flex-row items-center justify-between">
        <Text className="font-medium text-slate-800">
          {report.weekStartDate} – {report.weekEndDate}
        </Text>
        <Badge label={report.status} tone={STATUS_TONE[report.status]} />
      </View>
      <Text className="text-sm text-slate-600">{report.summary}</Text>
    </Card>
  );
}
