import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Text, View } from 'react-native';

import { Card } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { PrimaryButton } from '@/components/ui/primary-button';
import { ScreenContainer } from '@/components/ui/screen-container';
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
    </ScreenContainer>
  );
}
