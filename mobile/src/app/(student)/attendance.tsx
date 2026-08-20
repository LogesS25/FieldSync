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
import { attendanceSchema, type AttendanceFormValues } from '@/features/fieldwork/schemas';
import { ApiError } from '@/lib/api-client';
import * as attendanceService from '@/services/attendance';
import type { AttendanceRecord, VerificationStatus } from '@/types/fieldwork';

const STATUS_TONE: Record<VerificationStatus, BadgeTone> = {
  pending: 'warning',
  verified: 'success',
  rejected: 'danger',
};

export default function StudentAttendance() {
  const queryClient = useQueryClient();

  const { data: records, isLoading } = useQuery({
    queryKey: ['attendance'],
    queryFn: attendanceService.listAttendance,
  });

  const { data: summary } = useQuery({
    queryKey: ['attendance', 'summary'],
    queryFn: attendanceService.getAttendanceSummary,
  });

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<AttendanceFormValues>({
    resolver: zodResolver(attendanceSchema),
    defaultValues: { attendanceDate: '', hours: '' },
  });

  const mutation = useMutation({
    mutationFn: attendanceService.createAttendance,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['attendance'] });
      reset();
    },
  });

  const onSubmit = handleSubmit((values) =>
    mutation.mutate({ attendanceDate: values.attendanceDate, hours: Number(values.hours) }),
  );

  return (
    <ScreenContainer scroll>
      <Text className="mb-1 text-2xl font-bold text-slate-900">Attendance</Text>
      <Text className="mb-4 text-sm text-slate-500">Record your hours for each day.</Text>

      <Card className="mb-6 flex-row items-baseline justify-between bg-brand-600">
        <Text className="text-sm font-medium text-brand-100">Total Field Hours</Text>
        <Text className="text-2xl font-bold text-white">{summary ? summary.totalHours : '—'}</Text>
      </Card>

      <Card className="mb-6">
        <Controller
          control={control}
          name="attendanceDate"
          render={({ field: { onChange, value } }) => (
            <DateInput label="Date" value={value} onChange={onChange} error={errors.attendanceDate?.message} />
          )}
        />
        <Controller
          control={control}
          name="hours"
          render={({ field: { onChange, onBlur, value } }) => (
            <LabeledInput
              label="Hours"
              placeholder="e.g. 6.5"
              keyboardType="decimal-pad"
              onBlur={onBlur}
              onChangeText={onChange}
              value={value}
              error={errors.hours?.message}
            />
          )}
        />
        {mutation.isError ? (
          <Text className="mb-2 text-sm text-red-600">
            {mutation.error instanceof ApiError && mutation.error.status === 409
              ? 'Attendance for that date is already recorded, or you have no active practicum.'
              : 'Something went wrong. Please try again.'}
          </Text>
        ) : null}
        <PrimaryButton label={mutation.isPending ? 'Saving…' : 'Record Attendance'} onPress={onSubmit} disabled={mutation.isPending} />
      </Card>

      {isLoading ? null : !records || records.length === 0 ? (
        <EmptyState title="No attendance recorded yet" description="Days you record will show up here." />
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
    <Card className="flex-row items-center justify-between">
      <View>
        <Text className="font-medium text-slate-800">{record.attendanceDate}</Text>
        <Text className="text-sm text-slate-500">{record.hours} hours</Text>
      </View>
      <Badge label={record.verificationStatus} tone={STATUS_TONE[record.verificationStatus]} />
    </Card>
  );
}
