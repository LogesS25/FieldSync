import { zodResolver } from '@hookform/resolvers/zod';
import { Ionicons } from '@expo/vector-icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Controller, useForm } from 'react-hook-form';
import { Text, View } from 'react-native';

import { Badge, type BadgeTone } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { DateInput } from '@/components/ui/date-input';
import { EmptyState } from '@/components/ui/empty-state';
import { FormField } from '@/components/ui/form-field';
import { LabeledInput } from '@/components/ui/labeled-input';
import { LoadingState } from '@/components/ui/loading-state';
import { PageHeader } from '@/components/ui/page-header';
import { PillSelect } from '@/components/ui/pill-select';
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
      <PageHeader
        icon="people-circle-outline"
        title="Your Practicum Team"
        description="Pick an agency and both supervisors. Your team forms once both accept."
      />

      {requestsLoading ? (
        <LoadingState compact />
      ) : requests && requests.length > 0 ? (
        <View className="mb-6 gap-3">
          {requests.map((request) => (
            <TeamRequestCard key={request.id} request={request} />
          ))}
        </View>
      ) : null}

      {hasFormedTeam ? (
        <EmptyState
          icon="people-circle-outline"
          title="Your team is formed"
          description="You already have an active practicum team."
        />
      ) : (
        <Card>
          <FormField label="Agency" error={errors.agencyId?.message}>
            {agenciesLoading ? (
              <LoadingState compact />
            ) : (
              <Controller
                control={control}
                name="agencyId"
                render={({ field: { onChange, value } }) => (
                  <PillSelect
                    options={(agencies ?? []).map((agency) => ({ value: agency.id, label: agency.name }))}
                    value={value}
                    onChange={onChange}
                  />
                )}
              />
            )}
          </FormField>

          <FormField label="Faculty Supervisor" error={errors.facultySupervisorId?.message}>
            {facultyLoading ? (
              <LoadingState compact />
            ) : (
              <Controller
                control={control}
                name="facultySupervisorId"
                render={({ field: { onChange, value } }) => (
                  <PillSelect
                    options={(facultyList ?? []).map((faculty) => ({ value: faculty.id, label: faculty.fullName }))}
                    value={value}
                    onChange={onChange}
                  />
                )}
              />
            )}
          </FormField>

          <FormField
            label="Agency Supervisor"
            error={errors.agencySupervisorId?.message}
            hint={!selectedAgencyId ? 'Pick an agency first.' : undefined}
          >
            {!selectedAgencyId ? null : agencySupervisorsLoading ? (
              <LoadingState compact />
            ) : (
              <Controller
                control={control}
                name="agencySupervisorId"
                render={({ field: { onChange, value } }) => (
                  <PillSelect
                    options={(agencySupervisors ?? []).map((sup) => ({ value: sup.id, label: sup.fullName }))}
                    value={value}
                    onChange={onChange}
                  />
                )}
              />
            )}
          </FormField>

          <FormField label="Fieldwork Component" error={errors.fieldworkComponentId?.message}>
            {componentsLoading ? (
              <LoadingState compact />
            ) : (
              <Controller
                control={control}
                name="fieldworkComponentId"
                render={({ field: { onChange, value } }) => (
                  <PillSelect
                    options={(components ?? []).map((component) => ({ value: component.id, label: component.name }))}
                    value={value}
                    onChange={onChange}
                  />
                )}
              />
            )}
          </FormField>

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
            <View className="mb-4 flex-row items-center gap-2 rounded-xl bg-rose-50 px-4 py-3">
              <Ionicons name="alert-circle-outline" size={16} color="#e11d48" />
              <Text className="flex-1 text-sm text-rose-600">
                {mutation.error instanceof ApiError ? mutation.error.message : 'Something went wrong. Please try again.'}
              </Text>
            </View>
          ) : null}

          <PrimaryButton
            label="Send Team Request"
            onPress={onSubmit}
            disabled={mutation.isPending}
            loading={mutation.isPending}
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
      <Text className="mb-1 text-base font-bold text-slate-900">Weekly Feedback</Text>
      <Text className="mb-4 text-sm text-slate-500">Feedback from your supervisors, most recent first.</Text>

      {isLoading ? (
        <LoadingState compact />
      ) : !feedback || feedback.length === 0 ? (
        <EmptyState
          icon="chatbubble-ellipses-outline"
          title="No feedback yet"
          description="Your supervisors' weekly feedback will show up here."
        />
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
