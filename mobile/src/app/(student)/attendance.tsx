import { zodResolver } from '@hookform/resolvers/zod';
import { Ionicons } from '@expo/vector-icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Controller, useForm } from 'react-hook-form';
import { Text, View } from 'react-native';

import { Badge, type BadgeTone } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { DateInput } from '@/components/ui/date-input';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import { FormField } from '@/components/ui/form-field';
import { LabeledInput } from '@/components/ui/labeled-input';
import { LoadingState } from '@/components/ui/loading-state';
import { PageHeader } from '@/components/ui/page-header';
import { PillSelect } from '@/components/ui/pill-select';
import { PrimaryButton } from '@/components/ui/primary-button';
import { ScreenContainer } from '@/components/ui/screen-container';
import { attendanceSchema, type AttendanceFormValues } from '@/features/fieldwork/schemas';
import { ApiError } from '@/lib/api-client';
import * as attendanceService from '@/services/attendance';
import type { AttendanceRecord, AttendanceSession, ReviewDecision } from '@/types/fieldwork';

const STATUS_TONE: Record<ReviewDecision, BadgeTone> = {
  pending: 'warning',
  approved: 'success',
  rejected: 'danger',
};

const SESSION_OPTIONS: { value: AttendanceSession; label: string }[] = [
  { value: 'morning', label: 'Morning' },
  { value: 'evening', label: 'Evening' },
];

export default function StudentAttendance() {
  const queryClient = useQueryClient();

  const { data: records, isLoading, isError, refetch } = useQuery({
    queryKey: ['attendance'],
    queryFn: attendanceService.listAttendance,
  });

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<AttendanceFormValues>({
    resolver: zodResolver(attendanceSchema),
    defaultValues: { attendanceDate: '', session: 'morning', hours: '' },
  });

  const mutation = useMutation({
    mutationFn: attendanceService.createAttendance,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['attendance'] });
      reset();
    },
  });

  const onSubmit = handleSubmit((values) =>
    mutation.mutate({
      attendanceDate: values.attendanceDate,
      session: values.session,
      hours: values.hours ? Number(values.hours) : undefined,
    }),
  );

  return (
    <ScreenContainer scroll>
      <PageHeader
        icon="checkmark-done-circle-outline"
        title="Attendance"
        description="Record morning and evening attendance separately. Your agency supervisor reviews first, then your faculty supervisor."
      />

      <Card className="mb-6">
        <Controller
          control={control}
          name="attendanceDate"
          render={({ field: { onChange, value } }) => (
            <DateInput label="Date" value={value} onChange={onChange} error={errors.attendanceDate?.message} />
          )}
        />

        <FormField label="Session" error={errors.session?.message}>
          <Controller
            control={control}
            name="session"
            render={({ field: { onChange, value } }) => (
              <PillSelect options={SESSION_OPTIONS} value={value} onChange={onChange} />
            )}
          />
        </FormField>

        <Controller
          control={control}
          name="hours"
          render={({ field: { onChange, onBlur, value } }) => (
            <LabeledInput
              label="Hours (optional)"
              placeholder="e.g. 3.5"
              keyboardType="decimal-pad"
              onBlur={onBlur}
              onChangeText={onChange}
              value={value}
              error={errors.hours?.message}
            />
          )}
        />

        {mutation.isError ? (
          <View className="mb-4 flex-row items-center gap-2 rounded-xl bg-rose-50 px-4 py-3">
            <Ionicons name="alert-circle-outline" size={16} color="#e11d48" />
            <Text className="flex-1 text-sm text-rose-600">
              {mutation.error instanceof ApiError && mutation.error.status === 409
                ? 'Attendance for that date and session is already recorded, or you have no active practicum.'
                : 'Something went wrong. Please try again.'}
            </Text>
          </View>
        ) : null}
        <PrimaryButton
          label="Record Attendance"
          onPress={onSubmit}
          disabled={mutation.isPending}
          loading={mutation.isPending}
        />
      </Card>

      {isLoading ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState onRetry={() => refetch()} />
      ) : !records || records.length === 0 ? (
        <EmptyState
          icon="checkmark-done-circle-outline"
          title="No attendance recorded yet"
          description="Days you record will show up here."
        />
      ) : (
        <View className="gap-3">
          {records.map((record) => (
            <AttendanceCard key={record.id} record={record} />
          ))}
        </View>
      )}
    </ScreenContainer>
  );
}

function AttendanceCard({ record }: { record: AttendanceRecord }) {
  return (
    <Card>
      <View className="mb-2 flex-row items-center justify-between">
        <Text className="font-medium text-slate-800">
          {record.attendanceDate} · <Text className="capitalize">{record.session}</Text>
        </Text>
        {record.hours ? <Text className="text-sm text-slate-500">{record.hours}h</Text> : null}
      </View>
      <View className="flex-row gap-2">
        <Badge label={`Agency: ${record.agencyStatus}`} tone={STATUS_TONE[record.agencyStatus]} />
        <Badge label={`Faculty: ${record.facultyStatus}`} tone={STATUS_TONE[record.facultyStatus]} />
      </View>
    </Card>
  );
}
