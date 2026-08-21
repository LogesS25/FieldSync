import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Controller, useForm } from 'react-hook-form';
import { Text, View } from 'react-native';

import { Badge, type BadgeTone } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { LabeledInput } from '@/components/ui/labeled-input';
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

  const { data: report, isLoading } = useQuery({
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

  const onSubmit = handleSubmit((values) => mutation.mutate(values));

  return (
    <ScreenContainer scroll>
      <Text className="mb-1 text-2xl font-bold text-slate-900">Consolidated Report</Text>
      <Text className="mb-6 text-sm text-slate-500">
        One report covers your whole fieldwork period. It can&apos;t be edited once submitted, and your agency
        supervisor reviews it before your faculty supervisor does.
      </Text>

      {isLoading ? null : report ? (
        <Card>
          <View className="mb-3 flex-row gap-2">
            <Badge label={`Agency: ${report.agencyStatus}`} tone={STATUS_TONE[report.agencyStatus]} />
            <Badge label={`Faculty: ${report.facultyStatus}`} tone={STATUS_TONE[report.facultyStatus]} />
          </View>
          <Text className="text-sm text-slate-700">{report.summary}</Text>
        </Card>
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
            <Text className="mb-2 text-sm text-red-600">
              {mutation.error instanceof ApiError && mutation.error.status === 409
                ? 'A report has already been submitted, or you have no active practicum.'
                : 'Something went wrong. Please try again.'}
            </Text>
          ) : null}
          <PrimaryButton
            label={mutation.isPending ? 'Submitting…' : 'Submit Report'}
            onPress={onSubmit}
            disabled={mutation.isPending}
          />
        </Card>
      )}
    </ScreenContainer>
  );
}
