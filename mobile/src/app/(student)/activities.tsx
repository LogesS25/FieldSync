import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import * as DocumentPicker from 'expo-document-picker';
import { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { Pressable, Text, View } from 'react-native';

import { Badge, type BadgeTone } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { DateInput } from '@/components/ui/date-input';
import { EmptyState } from '@/components/ui/empty-state';
import { PrimaryButton } from '@/components/ui/primary-button';
import { ScreenContainer } from '@/components/ui/screen-container';
import { dailyReportSchema, type DailyReportFormValues } from '@/features/fieldwork/schemas';
import { ApiError } from '@/lib/api-client';
import { openAuthenticatedFile } from '@/lib/open-file';
import * as dailyReportsService from '@/services/dailyreports';
import type { PickedFile } from '@/services/dailyreports';
import { useAuthStore } from '@/stores/auth-store';
import type { DailyReport } from '@/types/dailyreport';
import type { ReviewDecision } from '@/types/fieldwork';

const STATUS_TONE: Record<ReviewDecision, BadgeTone> = {
  pending: 'warning',
  approved: 'success',
  rejected: 'danger',
};

export default function StudentDailyReports() {
  const queryClient = useQueryClient();
  const accessToken = useAuthStore((state) => state.accessToken);
  const [pickedFile, setPickedFile] = useState<PickedFile | null>(null);
  const [pickerError, setPickerError] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['daily-reports', 'mine'],
    queryFn: dailyReportsService.listMyDailyReports,
  });

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<DailyReportFormValues>({
    resolver: zodResolver(dailyReportSchema),
    defaultValues: { reportDate: '' },
  });

  const mutation = useMutation({
    mutationFn: (values: DailyReportFormValues) => {
      if (!pickedFile) throw new Error('no file picked');
      return dailyReportsService.submitDailyReport({ reportDate: values.reportDate, file: pickedFile });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['daily-reports', 'mine'] });
      reset({ reportDate: '' });
      setPickedFile(null);
    },
  });

  const pickFile = async () => {
    setPickerError(null);
    const result = await DocumentPicker.getDocumentAsync({ type: 'application/pdf' });
    if (result.canceled || result.assets.length === 0) return;
    const asset = result.assets[0];
    setPickedFile({ uri: asset.uri, name: asset.name, mimeType: asset.mimeType ?? 'application/pdf' });
  };

  const onSubmit = handleSubmit((values) => {
    if (!pickedFile) {
      setPickerError('Choose a PDF first');
      return;
    }
    mutation.mutate(values);
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
        Upload your handwritten fieldwork report for the day as a PDF. Your agency supervisor reviews it before your
        faculty supervisor does.
      </Text>

      <Card className="mb-6">
        <Controller
          control={control}
          name="reportDate"
          render={({ field: { onChange, value } }) => (
            <DateInput label="Report date" value={value} onChange={onChange} error={errors.reportDate?.message} />
          )}
        />

        <Text className="mb-1 text-sm font-medium text-slate-700">Report file</Text>
        <Pressable
          onPress={pickFile}
          className="mb-1 rounded-xl border border-dashed border-slate-300 bg-slate-50 px-4 py-3"
        >
          <Text className="text-sm text-slate-600">{pickedFile ? pickedFile.name : 'Choose a PDF…'}</Text>
        </Pressable>
        {pickerError ? <Text className="mb-2 text-sm text-red-600">{pickerError}</Text> : null}

        {mutation.isError ? (
          <Text className="mb-2 mt-2 text-sm text-red-600">
            {mutation.error instanceof ApiError ? mutation.error.message : 'Something went wrong. Please try again.'}
          </Text>
        ) : null}

        <PrimaryButton
          label={mutation.isPending ? 'Submitting…' : 'Submit Report'}
          onPress={onSubmit}
          disabled={mutation.isPending}
        />
      </Card>

      {isLoading ? null : !data || data.length === 0 ? (
        <EmptyState title="No reports submitted yet" description="Reports you submit will show up here." />
      ) : (
        <View className="gap-3">
          {data.map((report) => (
            <ReportCard key={report.id} report={report} onView={() => viewFile(report)} />
          ))}
        </View>
      )}
    </ScreenContainer>
  );
}

function ReportCard({ report, onView }: { report: DailyReport; onView: () => void }) {
  return (
    <Card>
      <View className="mb-2 flex-row items-center justify-between">
        <Text className="font-medium text-slate-800">{report.reportDate}</Text>
        <View className="flex-row gap-2">
          <Badge label={`Agency: ${report.agencyStatus}`} tone={STATUS_TONE[report.agencyStatus]} />
          <Badge label={`Faculty: ${report.facultyStatus}`} tone={STATUS_TONE[report.facultyStatus]} />
        </View>
      </View>
      <Pressable onPress={onView}>
        <Text className="text-sm text-brand-600">{report.filename}</Text>
      </Pressable>
    </Card>
  );
}
