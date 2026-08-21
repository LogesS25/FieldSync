import { useQuery } from '@tanstack/react-query';
import { FlatList, Text, View } from 'react-native';

import { Avatar } from '@/components/ui/avatar';
import { Badge, type BadgeTone } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import { LoadingState } from '@/components/ui/loading-state';
import { PageHeader } from '@/components/ui/page-header';
import { ScreenContainer } from '@/components/ui/screen-container';
import * as practicumService from '@/services/practicums';
import type { AssignedStudent, PracticumStatus } from '@/types/practicum';

const STATUS_TONE: Record<PracticumStatus, BadgeTone> = {
  active: 'success',
  completed: 'neutral',
  terminated: 'danger',
};

export default function SupervisorStudents() {
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['students'],
    queryFn: practicumService.listMyStudents,
  });

  return (
    <ScreenContainer>
      <PageHeader
        icon="people-outline"
        title="Your Students"
        description={data ? `${data.length} student${data.length === 1 ? '' : 's'} assigned to you` : ' '}
      />

      {isLoading ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState onRetry={() => refetch()} />
      ) : !data || data.length === 0 ? (
        <EmptyState
          icon="people-outline"
          title="No students assigned yet"
          description="Students will appear here once an administrator assigns you to their practicum."
        />
      ) : (
        <FlatList
          data={data}
          keyExtractor={(item) => item.practicumId}
          renderItem={({ item }) => <StudentCard student={item} />}
          ItemSeparatorComponent={() => <View className="h-3" />}
        />
      )}
    </ScreenContainer>
  );
}

function StudentCard({ student }: { student: AssignedStudent }) {
  return (
    <Card>
      <View className="flex-row items-center gap-3">
        <Avatar name={student.studentName} />
        <View className="flex-1">
          <Text className="font-semibold text-slate-900">{student.studentName}</Text>
          <Text className="text-xs text-slate-500">{student.studentEmail}</Text>
        </View>
        <Badge label={student.practicumStatus} tone={STATUS_TONE[student.practicumStatus]} />
      </View>

      <View className="mt-4 flex-row justify-between border-t border-slate-100 pt-3">
        <Text className="text-xs text-slate-500">{student.institutionName}</Text>
        <Text className="text-xs text-slate-500">{student.agencyName ?? 'Not yet placed'}</Text>
      </View>
    </Card>
  );
}
