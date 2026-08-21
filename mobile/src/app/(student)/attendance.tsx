import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Controller, useForm } from 'react-hook-form';
import { Pressable, Text, View } from 'react-native';

import { Badge, type BadgeTone } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { DateInput } from '@/components/ui/date-input';
import { EmptyState } from '@/components/ui/empty-state';
import { LabeledInput } from '@/components/ui/labeled-input';
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

  const { data: records, isLoading } = useQuery({
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
      <Text className="mb-1 text-2xl font-bold text-slate-900">Attendance</Text>
      <Text className="mb-6 text-sm text-slate-500">
        Record morning and evening attendance separately. Your agency supervisor reviews first, then your faculty
        supervisor.
      </Text>

      <Card className="mb-6">
        <Controller
          control={control}
          name="attendanceDate"
          render={({ field: { onChange, value } }) => (
            <DateInput label="Date" value={value} onChange={onChange} error={errors.attendanceDate?.message} />
          )}
        />

        <Text className="mb-1 text-sm font-medium text-slate-700">Session</Text>
        <Controller
          control={control}
          name="session"
          render={({ field: { onChange, value } }) => (
            <View className="mb-3 flex-row gap-2">
              {SESSION_OPTIONS.map((option) => (
                <Pressable
                  key={option.value}
                  onPress={() => onChange(option.value)}
                  className={`rounded-full border px-4 py-2 ${
                    value === option.value ? 'border-brand-600 bg-brand-600' : 'border-slate-300 bg-white'
                  }`}
                >
                  <Text className={value === option.value ? 'text-white' : 'text-slate-700'}>{option.label}</Text>
                </Pressable>
              ))}
            </View>
          )}
        />

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
          <Text className="mb-2 text-sm text-red-600">
            {mutation.error instanceof ApiError && mutation.error.status === 409
              ? 'Attendance for that date and session is already recorded, or you have no active practicum.'
              : 'Something went wrong. Please try again.'}
          </Text>
        ) : null}
        <PrimaryButton
          label={mutation.isPending ? 'Saving…' : 'Record Attendance'}
          onPress={onSubmit}
          disabled={mutation.isPending}
        />
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
