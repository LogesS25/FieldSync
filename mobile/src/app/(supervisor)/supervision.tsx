import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { Pressable, Text, View } from 'react-native';

import { Card } from '@/components/ui/card';
import { DateInput } from '@/components/ui/date-input';
import { EmptyState } from '@/components/ui/empty-state';
import { LabeledInput } from '@/components/ui/labeled-input';
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

  const { data, isLoading } = useQuery({
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
      <Text className="mb-1 text-2xl font-bold text-slate-900">Team Requests</Text>
      <Text className="mb-6 text-sm text-slate-500">Students requesting you as a supervisor.</Text>

      {isLoading ? null : pending.length === 0 ? (
        <EmptyState title="No pending requests" description="Requests naming you as a supervisor will show up here." />
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
      <Text className="mb-1 text-base font-semibold text-slate-900">Weekly Feedback</Text>
      <Text className="mb-4 text-sm text-slate-500">
        Give each student feedback for the week. This is required every weekend.
      </Text>

      <Text className="mb-1 text-sm font-medium text-slate-700">Student</Text>
      {studentsLoading ? null : !students || students.length === 0 ? (
        <Text className="mb-2 text-sm text-slate-400">No students assigned yet.</Text>
      ) : (
        <View className="mb-3 flex-row flex-wrap gap-2">
          {students.map((student) => (
            <Pressable
              key={student.practicumId}
              onPress={() => setSelectedPracticumId(student.practicumId)}
              className={`rounded-full border px-4 py-2 ${
                selectedPracticumId === student.practicumId
                  ? 'border-brand-600 bg-brand-600'
                  : 'border-slate-300 bg-white'
              }`}
            >
              <Text className={selectedPracticumId === student.practicumId ? 'text-white' : 'text-slate-700'}>
                {student.studentName}
              </Text>
            </Pressable>
          ))}
        </View>
      )}

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

      {!selectedPracticumId ? (
        <Text className="mb-2 text-sm text-slate-400">Pick a student first.</Text>
      ) : null}
      {mutation.isError ? (
        <Text className="mb-2 text-sm text-red-600">
          {mutation.error instanceof ApiError ? mutation.error.message : 'Something went wrong. Please try again.'}
        </Text>
      ) : null}
      {mutation.isSuccess ? <Text className="mb-2 text-sm text-emerald-600">Feedback submitted.</Text> : null}

      <PrimaryButton
        label={mutation.isPending ? 'Submitting…' : 'Submit Feedback'}
        onPress={onSubmit}
        disabled={mutation.isPending || !selectedPracticumId}
      />
    </Card>
  );
}
