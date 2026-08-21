import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Controller, useForm } from 'react-hook-form';
import { ActivityIndicator, Pressable, Text, View } from 'react-native';

import { Badge, type BadgeTone } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { DateInput } from '@/components/ui/date-input';
import { EmptyState } from '@/components/ui/empty-state';
import { LabeledInput } from '@/components/ui/labeled-input';
import { PrimaryButton } from '@/components/ui/primary-button';
import { ScreenContainer } from '@/components/ui/screen-container';
import { teamRequestSchema, type TeamRequestFormValues } from '@/features/fieldwork/schemas';
import { ApiError } from '@/lib/api-client';
import * as directoryService from '@/services/directory';
import * as feedbackService from '@/services/feedback';
import * as teamRequestsService from '@/services/teamrequests';
import type { TeamRequest, TeamRequestDecision } from '@/types/fieldwork';
import type { WeeklyFeedback } from '@/types/feedback';

const DECISION_TONE: Record<TeamRequestDecision, BadgeTone> = {
  pending: 'warning',
  accepted: 'success',
  rejected: 'danger',
};

export default function StudentTeam() {
  const queryClient = useQueryClient();

  const { data: requests, isLoading: requestsLoading } = useQuery({
    queryKey: ['team-requests', 'me'],
    queryFn: teamRequestsService.listMyTeamRequests,
  });

  const hasFormedTeam = requests?.some((r) => r.formedPracticumId);

  const { data: agencies, isLoading: agenciesLoading } = useQuery({
    queryKey: ['agencies', 'mine'],
    queryFn: directoryService.listMyAgencies,
    enabled: !hasFormedTeam,
  });
  const { data: facultyList, isLoading: facultyLoading } = useQuery({
    queryKey: ['faculty-supervisors', 'mine'],
    queryFn: teamRequestsService.listMyFacultySupervisors,
    enabled: !hasFormedTeam,
  });
  const { data: components, isLoading: componentsLoading } = useQuery({
    queryKey: ['fieldwork-components', 'mine'],
    queryFn: directoryService.listMyFieldworkComponents,
    enabled: !hasFormedTeam,
  });

  const {
    control,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<TeamRequestFormValues>({
    resolver: zodResolver(teamRequestSchema),
    defaultValues: {
      agencyId: '',
      facultySupervisorId: '',
      agencySupervisorId: '',
      fieldworkComponentId: '',
      fieldworkDescription: '',
      startDate: '',
    },
  });

  const selectedAgencyId = watch('agencyId');
  const { data: agencySupervisors, isLoading: agencySupervisorsLoading } = useQuery({
    queryKey: ['agency-supervisors', selectedAgencyId],
    queryFn: () => teamRequestsService.listAgencySupervisors(selectedAgencyId),
    enabled: !!selectedAgencyId,
  });

  const mutation = useMutation({
    mutationFn: teamRequestsService.createTeamRequest,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['team-requests', 'me'] });
    },
  });

  const onSubmit = handleSubmit((values) => mutation.mutate(values));

  return (
    <ScreenContainer scroll>
      <Text className="mb-1 text-2xl font-bold text-slate-900">Your Practicum Team</Text>
      <Text className="mb-6 text-sm text-slate-500">
        Pick an agency and both supervisors. Your team forms once both accept.
      </Text>

      {requestsLoading ? (
        <ActivityIndicator className="mb-6" color="#4f46e5" />
      ) : requests && requests.length > 0 ? (
        <View className="mb-6 gap-3">
          {requests.map((request) => (
            <TeamRequestCard key={request.id} request={request} />
          ))}
        </View>
      ) : null}

      {hasFormedTeam ? (
        <EmptyState title="Your team is formed" description="You already have an active practicum team." />
      ) : (
        <Card>
          <Text className="mb-1 text-sm font-medium text-slate-700">Agency</Text>
          {agenciesLoading ? (
            <ActivityIndicator />
          ) : (
            <Controller
              control={control}
              name="agencyId"
              render={({ field: { onChange, value } }) => (
                <View className="mb-1 flex-row flex-wrap gap-2">
                  {(agencies ?? []).map((agency) => (
                    <Pressable
                      key={agency.id}
                      onPress={() => onChange(agency.id)}
                      className={`rounded-full border px-4 py-2 ${
                        value === agency.id ? 'border-brand-600 bg-brand-600' : 'border-slate-300 bg-white'
                      }`}
                    >
                      <Text className={value === agency.id ? 'text-white' : 'text-slate-700'}>{agency.name}</Text>
                    </Pressable>
                  ))}
                </View>
              )}
            />
          )}
          {errors.agencyId ? <Text className="mb-2 text-sm text-red-600">{errors.agencyId.message}</Text> : null}

          <Text className="mb-1 mt-2 text-sm font-medium text-slate-700">Faculty Supervisor</Text>
          {facultyLoading ? (
            <ActivityIndicator />
          ) : (
            <Controller
              control={control}
              name="facultySupervisorId"
              render={({ field: { onChange, value } }) => (
                <View className="mb-1 flex-row flex-wrap gap-2">
                  {(facultyList ?? []).map((faculty) => (
                    <Pressable
                      key={faculty.id}
                      onPress={() => onChange(faculty.id)}
                      className={`rounded-full border px-4 py-2 ${
                        value === faculty.id ? 'border-brand-600 bg-brand-600' : 'border-slate-300 bg-white'
                      }`}
                    >
                      <Text className={value === faculty.id ? 'text-white' : 'text-slate-700'}>{faculty.fullName}</Text>
                    </Pressable>
                  ))}
                </View>
              )}
            />
          )}
          {errors.facultySupervisorId ? (
            <Text className="mb-2 text-sm text-red-600">{errors.facultySupervisorId.message}</Text>
          ) : null}

          <Text className="mb-1 mt-2 text-sm font-medium text-slate-700">Agency Supervisor</Text>
          {!selectedAgencyId ? (
            <Text className="mb-2 text-sm text-slate-400">Pick an agency first.</Text>
          ) : agencySupervisorsLoading ? (
            <ActivityIndicator />
          ) : (
            <Controller
              control={control}
              name="agencySupervisorId"
              render={({ field: { onChange, value } }) => (
                <View className="mb-1 flex-row flex-wrap gap-2">
                  {(agencySupervisors ?? []).map((sup) => (
                    <Pressable
                      key={sup.id}
                      onPress={() => onChange(sup.id)}
                      className={`rounded-full border px-4 py-2 ${
                        value === sup.id ? 'border-brand-600 bg-brand-600' : 'border-slate-300 bg-white'
                      }`}
                    >
                      <Text className={value === sup.id ? 'text-white' : 'text-slate-700'}>{sup.fullName}</Text>
                    </Pressable>
                  ))}
                </View>
              )}
            />
          )}
          {errors.agencySupervisorId ? (
            <Text className="mb-2 text-sm text-red-600">{errors.agencySupervisorId.message}</Text>
          ) : null}

          <Text className="mb-1 mt-2 text-sm font-medium text-slate-700">Fieldwork Component</Text>
          {componentsLoading ? (
            <ActivityIndicator />
          ) : (
            <Controller
              control={control}
              name="fieldworkComponentId"
              render={({ field: { onChange, value } }) => (
                <View className="mb-1 flex-row flex-wrap gap-2">
                  {(components ?? []).map((component) => (
                    <Pressable
                      key={component.id}
                      onPress={() => onChange(component.id)}
                      className={`rounded-full border px-4 py-2 ${
                        value === component.id ? 'border-brand-600 bg-brand-600' : 'border-slate-300 bg-white'
                      }`}
                    >
                      <Text className={value === component.id ? 'text-white' : 'text-slate-700'}>{component.name}</Text>
                    </Pressable>
                  ))}
                </View>
              )}
            />
          )}
          {errors.fieldworkComponentId ? (
            <Text className="mb-2 text-sm text-red-600">{errors.fieldworkComponentId.message}</Text>
          ) : null}

          <Controller
            control={control}
            name="fieldworkDescription"
            render={({ field: { onChange, onBlur, value } }) => (
              <LabeledInput
                label="Fieldwork"
                placeholder="Describe the fieldwork you'll be doing"
                multiline
                numberOfLines={3}
                onBlur={onBlur}
                onChangeText={onChange}
                value={value}
                error={errors.fieldworkDescription?.message}
              />
            )}
          />

          <Controller
            control={control}
            name="startDate"
            render={({ field: { onChange, value } }) => (
              <DateInput label="Start date" value={value} onChange={onChange} error={errors.startDate?.message} />
            )}
          />

          {mutation.isError ? (
            <Text className="mb-2 text-sm text-red-600">
              {mutation.error instanceof ApiError ? mutation.error.message : 'Something went wrong. Please try again.'}
            </Text>
          ) : null}

          <PrimaryButton
            label={mutation.isPending ? 'Sending…' : 'Send Team Request'}
            onPress={onSubmit}
            disabled={mutation.isPending}
          />
        </Card>
      )}

      <View className="mt-8">
        <FeedbackReceived />
      </View>
    </ScreenContainer>
  );
}

function FeedbackReceived() {
  const { data: feedback, isLoading } = useQuery({
    queryKey: ['feedback', 'me'],
    queryFn: feedbackService.listFeedbackForMe,
  });

  return (
    <View>
      <Text className="mb-1 text-lg font-bold text-slate-900">Weekly Feedback</Text>
      <Text className="mb-4 text-sm text-slate-500">Feedback from your supervisors, most recent first.</Text>

      {isLoading ? (
        <ActivityIndicator color="#4f46e5" />
      ) : !feedback || feedback.length === 0 ? (
        <EmptyState title="No feedback yet" description="Your supervisors' weekly feedback will show up here." />
      ) : (
        <View className="gap-3">
          {[...feedback]
            .sort((a, b) => b.weekStartDate.localeCompare(a.weekStartDate))
            .map((entry) => (
              <FeedbackCard key={entry.id} entry={entry} />
            ))}
        </View>
      )}
    </View>
  );
}

function FeedbackCard({ entry }: { entry: WeeklyFeedback }) {
  return (
    <Card>
      <Text className="mb-2 text-xs font-medium text-slate-500">Week of {entry.weekStartDate}</Text>
      <Text className="text-sm text-slate-700">{entry.feedback}</Text>
    </Card>
  );
}

function TeamRequestCard({ request }: { request: TeamRequest }) {
  return (
    <Card>
      <Text className="mb-2 text-sm text-slate-600">{request.fieldworkDescription}</Text>
      <View className="flex-row gap-2">
        <Badge label={`Faculty: ${request.facultyDecision}`} tone={DECISION_TONE[request.facultyDecision]} />
        <Badge label={`Agency: ${request.agencyDecision}`} tone={DECISION_TONE[request.agencyDecision]} />
      </View>
    </Card>
  );
}
