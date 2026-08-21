import { zodResolver } from '@hookform/resolvers/zod';
import { Ionicons } from '@expo/vector-icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { Text, View } from 'react-native';

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
import { feedbackSchema, type FeedbackFormValues } from '@/features/fieldwork/schemas';
import { ApiError } from '@/lib/api-client';
import * as feedbackService from '@/services/feedback';
import * as practicumService from '@/services/practicums';
import * as teamRequestsService from '@/services/teamrequests';
import { useAuthStore } from '@/stores/auth-store';

export default function SupervisorTeamRequests() {
  const queryClient = useQueryClient();
  const user = useAuthStore((state) => state.user);

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['team-requests', 'pending'],
    queryFn: teamRequestsService.listPendingTeamRequests,
  });

  const mutation = useMutation({
    mutationFn: ({ id, decision }: { id: string; decision: 'accepted' | 'rejected' }) =>
      teamRequestsService.respondToTeamRequest(id, decision),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['team-requests', 'pending'] }),
  });

  // A request is "yours to decide" if your side of it is still pending —
  // once you've responded it drops off this screen even if the other
  // supervisor hasn't decided yet.
  const pending = (data ?? []).filter((r) =>
    user?.role === 'faculty_supervisor' ? r.facultyDecision === 'pending' : r.agencyDecision === 'pending',
  );

  return (
    <ScreenContainer scroll>
      <PageHeader
        icon="people-circle-outline"
        title="Team Requests"
        description="Students requesting you as a supervisor."
      />

      {isLoading ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState onRetry={() => refetch()} />
      ) : pending.length === 0 ? (
        <EmptyState
          icon="people-circle-outline"
          title="No pending requests"
          description="Requests naming you as a supervisor will show up here."
        />
      ) : (
        <View className="gap-3">
          {pending.map((request) => (
            <Card key={request.id}>
              <Text className="mb-3 text-sm text-slate-700">{request.fieldworkDescription}</Text>
              <Text className="mb-3 text-xs text-slate-500">Start date: {request.startDate}</Text>
              <View className="flex-row gap-3">
                <View className="flex-1">
                  <PrimaryButton
                    label="Accept"
                    onPress={() => mutation.mutate({ id: request.id, decision: 'accepted' })}
                    disabled={mutation.isPending}
                  />
                </View>
                <View className="flex-1">
                  <PrimaryButton
                    label="Reject"
                    variant="danger"
                    onPress={() => mutation.mutate({ id: request.id, decision: 'rejected' })}
                    disabled={mutation.isPending}
                  />
                </View>
              </View>
            </Card>
          ))}
        </View>
      )}

      <View className="mt-8">
        <WeeklyFeedbackForm />
      </View>
    </ScreenContainer>
  );
}

function WeeklyFeedbackForm() {
  const queryClient = useQueryClient();
  const [selectedPracticumId, setSelectedPracticumId] = useState<string | null>(null);

  const { data: students, isLoading: studentsLoading } = useQuery({
    queryKey: ['students'],
    queryFn: practicumService.listMyStudents,
  });

  const {
    control,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FeedbackFormValues>({
    resolver: zodResolver(feedbackSchema),
    defaultValues: { weekStartDate: '', feedback: '' },
  });

  const mutation = useMutation({
    mutationFn: (values: FeedbackFormValues) =>
      feedbackService.submitFeedback({ practicumId: selectedPracticumId!, ...values }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feedback', 'mine'] });
      reset({ weekStartDate: '', feedback: '' });
    },
  });

  const onSubmit = handleSubmit((values) => {
    if (!selectedPracticumId) return;
    mutation.mutate(values);
  });

  return (
    <Card>
      <Text className="mb-1 text-base font-bold text-slate-900">Weekly Feedback</Text>
      <Text className="mb-4 text-sm text-slate-500">
        Give each student feedback for the week. This is required every weekend.
      </Text>

      <FormField label="Student" hint={!studentsLoading && (!students || students.length === 0) ? 'No students assigned yet.' : undefined}>
        {studentsLoading ? (
          <LoadingState compact />
        ) : students && students.length > 0 ? (
          <PillSelect
            options={students.map((student) => ({ value: student.practicumId, label: student.studentName }))}
            value={selectedPracticumId ?? ''}
            onChange={setSelectedPracticumId}
          />
        ) : null}
      </FormField>

      <Controller
        control={control}
        name="weekStartDate"
        render={({ field: { onChange, value } }) => (
          <DateInput label="Week start date" value={value} onChange={onChange} error={errors.weekStartDate?.message} />
        )}
      />

      <Controller
        control={control}
        name="feedback"
        render={({ field: { onChange, onBlur, value } }) => (
          <LabeledInput
            label="Feedback"
            placeholder="How did the student do this week?"
            multiline
            numberOfLines={4}
            onBlur={onBlur}
            onChangeText={onChange}
            value={value}
            error={errors.feedback?.message}
          />
        )}
      />

      {!selectedPracticumId ? <Text className="mb-3 text-sm text-slate-400">Pick a student first.</Text> : null}
      {mutation.isError ? (
        <View className="mb-4 flex-row items-center gap-2 rounded-xl bg-rose-50 px-4 py-3">
          <Ionicons name="alert-circle-outline" size={16} color="#e11d48" />
          <Text className="flex-1 text-sm text-rose-600">
            {mutation.error instanceof ApiError ? mutation.error.message : 'Something went wrong. Please try again.'}
          </Text>
        </View>
      ) : null}
      {mutation.isSuccess ? (
        <View className="mb-4 flex-row items-center gap-2 rounded-xl bg-emerald-50 px-4 py-3">
          <Ionicons name="checkmark-circle-outline" size={16} color="#059669" />
          <Text className="text-sm font-medium text-emerald-700">Feedback submitted.</Text>
        </View>
      ) : null}

      <PrimaryButton
        label="Submit Feedback"
        onPress={onSubmit}
        disabled={mutation.isPending || !selectedPracticumId}
        loading={mutation.isPending}
      />
    </Card>
  );
}
