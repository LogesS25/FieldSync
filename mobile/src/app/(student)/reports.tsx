import { zodResolver } from '@hookform/resolvers/zod';
import { Ionicons } from '@expo/vector-icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Controller, useForm } from 'react-hook-form';
import { Text, View } from 'react-native';

import { Badge, type BadgeTone } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { ErrorState } from '@/components/ui/error-state';
import { LabeledInput } from '@/components/ui/labeled-input';
import { LoadingState } from '@/components/ui/loading-state';
import { PageHeader } from '@/components/ui/page-header';
import { PrimaryButton } from '@/components/ui/primary-button';
import { ScreenContainer } from '@/components/ui/screen-container';
import { consolidatedReportSchema, type ConsolidatedReportFormValues } from '@/features/fieldwork/schemas';
import { ApiError } from '@/lib/api-client';
import * as reportsService from '@/services/reports';
import type { ReviewDecision } from '@/types/fieldwork';

const STATUS_TONE: Record<ReviewDecision, BadgeTone> = {
  pending: 'warning',
  approved: 'success',
  rejected: 'danger',
};

export default function StudentReports() {
  const queryClient = useQueryClient();

  const { data: report, isLoading, isError, refetch } = useQuery({
    queryKey: ['consolidated-report', 'me'],
    queryFn: reportsService.getMyConsolidatedReport,
  });

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<ConsolidatedReportFormValues>({
    resolver: zodResolver(consolidatedReportSchema),
    defaultValues: { summary: '' },
  });

  const mutation = useMutation({
    mutationFn: (values: ConsolidatedReportFormValues) => reportsService.submitConsolidatedReport(values.summary),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['consolidated-report', 'me'] });
    },
  });

  const resubmitMutation = useMutation({
    mutationFn: (values: ConsolidatedReportFormValues) =>
      reportsService.resubmitConsolidatedReport(report!.id, values.summary),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['consolidated-report', 'me'] });
    },
  });

  const onSubmit = handleSubmit((values) => mutation.mutate(values));
  const onResubmit = handleSubmit((values) => resubmitMutation.mutate(values));

  const isRejected = report && (report.agencyStatus === 'rejected' || report.facultyStatus === 'rejected');

  return (
    <ScreenContainer scroll>
      <PageHeader
        icon="bar-chart-outline"
        title="Consolidated Report"
        description="One report covers your whole fieldwork period. It can't be edited once submitted, and your agency supervisor reviews it before your faculty supervisor does."
      />

      {isLoading ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState onRetry={() => refetch()} />
      ) : report ? (
        <View className="gap-4">
          <Card>
            <View className="mb-3 flex-row gap-2">
              <Badge label={`Agency: ${report.agencyStatus}`} tone={STATUS_TONE[report.agencyStatus]} />
              <Badge label={`Faculty: ${report.facultyStatus}`} tone={STATUS_TONE[report.facultyStatus]} />
            </View>
            <Text className="text-sm text-slate-700">{report.summary}</Text>
          </Card>

          {isRejected ? (
            <Card>
              <Text className="mb-1 text-base font-semibold text-slate-900">Resubmit Report</Text>
              <Text className="mb-3 text-sm text-slate-500">
                Your report was rejected. Update your summary and resubmit for review.
              </Text>
              <Controller
                control={control}
                name="summary"
                render={({ field: { onChange, onBlur, value } }) => (
                  <LabeledInput
                    label="Summary"
                    placeholder="Summarize your fieldwork for the whole period…"
                    multiline
                    numberOfLines={6}
                    onBlur={onBlur}
                    onChangeText={onChange}
                    value={value}
                    error={errors.summary?.message}
                  />
                )}
              />
              {resubmitMutation.isError ? (
                <View className="mb-4 flex-row items-center gap-2 rounded-xl bg-rose-50 px-4 py-3">
                  <Ionicons name="alert-circle-outline" size={16} color="#e11d48" />
                  <Text className="flex-1 text-sm text-rose-600">
                    {resubmitMutation.error instanceof ApiError
                      ? resubmitMutation.error.message
                      : 'Something went wrong. Please try again.'}
                  </Text>
                </View>
              ) : null}
              <PrimaryButton
                label="Resubmit Report"
                onPress={onResubmit}
                disabled={resubmitMutation.isPending}
                loading={resubmitMutation.isPending}
              />
            </Card>
          ) : null}
        </View>
      ) : (
        <Card>
          <Controller
            control={control}
            name="summary"
            render={({ field: { onChange, onBlur, value } }) => (
              <LabeledInput
                label="Summary"
                placeholder="Summarize your fieldwork for the whole period…"
                multiline
                numberOfLines={6}
                onBlur={onBlur}
                onChangeText={onChange}
                value={value}
                error={errors.summary?.message}
              />
            )}
          />
          {mutation.isError ? (
            <View className="mb-4 flex-row items-center gap-2 rounded-xl bg-rose-50 px-4 py-3">
              <Ionicons name="alert-circle-outline" size={16} color="#e11d48" />
              <Text className="flex-1 text-sm text-rose-600">
                {mutation.error instanceof ApiError && mutation.error.status === 409
                  ? 'A report has already been submitted, or you have no active practicum.'
                  : 'Something went wrong. Please try again.'}
              </Text>
            </View>
          ) : null}
          <PrimaryButton
            label="Submit Report"
            onPress={onSubmit}
            disabled={mutation.isPending}
            loading={mutation.isPending}
          />
        </Card>
      )}
    </ScreenContainer>
  );
}
