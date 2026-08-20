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
import { fieldActivitySchema, type FieldActivityFormValues } from '@/features/fieldwork/schemas';
import { ApiError } from '@/lib/api-client';
import * as activitiesService from '@/services/activities';
import type { FieldActivity, VerificationStatus } from '@/types/fieldwork';

const STATUS_TONE: Record<VerificationStatus, BadgeTone> = {
  pending: 'warning',
  verified: 'success',
  rejected: 'danger',
};

export default function StudentActivities() {
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['field-activities'],
    queryFn: activitiesService.listFieldActivities,
  });

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FieldActivityFormValues>({
    resolver: zodResolver(fieldActivitySchema),
    defaultValues: { activityDate: '', description: '' },
  });

  const mutation = useMutation({
    mutationFn: activitiesService.createFieldActivity,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['field-activities'] });
      reset();
    },
  });

  const onSubmit = handleSubmit((values) => mutation.mutate(values));

  return (
    <ScreenContainer scroll>
      <Text className="mb-1 text-2xl font-bold text-slate-900">Field Activities</Text>
      <Text className="mb-6 text-sm text-slate-500">Log what you worked on each day.</Text>

      <Card className="mb-6">
        <Controller
          control={control}
          name="activityDate"
          render={({ field: { onChange, value } }) => (
            <DateInput label="Date" value={value} onChange={onChange} error={errors.activityDate?.message} />
          )}
        />
        <Controller
          control={control}
          name="description"
          render={({ field: { onChange, onBlur, value } }) => (
            <LabeledInput
              label="Description"
              placeholder="What did you work on?"
              multiline
              numberOfLines={3}
              onBlur={onBlur}
              onChangeText={onChange}
              value={value}
              error={errors.description?.message}
            />
          )}
        />
        {mutation.isError ? (
          <Text className="mb-2 text-sm text-red-600">
            {mutation.error instanceof ApiError && mutation.error.status === 409
              ? 'You need an active practicum before logging activities.'
              : 'Something went wrong. Please try again.'}
          </Text>
        ) : null}
        <PrimaryButton label={mutation.isPending ? 'Saving…' : 'Log Activity'} onPress={onSubmit} disabled={mutation.isPending} />
      </Card>

      {isLoading ? null : !data || data.length === 0 ? (
        <EmptyState title="No activities logged yet" description="Entries you log will show up here." />
      ) : (
        <View className="gap-3">
          {data.map((activity) => (
            <ActivityCard key={activity.id} activity={activity} />
          ))}
        </View>
      )}
    </ScreenContainer>
  );
}

function ActivityCard({ activity }: { activity: FieldActivity }) {
  return (
    <Card>
      <View className="mb-2 flex-row items-center justify-between">
        <Text className="font-medium text-slate-800">{activity.activityDate}</Text>
        <Badge label={activity.verificationStatus} tone={STATUS_TONE[activity.verificationStatus]} />
      </View>
      <Text className="text-sm text-slate-600">{activity.description}</Text>
    </Card>
  );
}
