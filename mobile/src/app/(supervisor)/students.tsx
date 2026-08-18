import { useQuery } from '@tanstack/react-query';
import { ActivityIndicator, FlatList, Text, View } from 'react-native';

import { Avatar } from '@/components/ui/avatar';
import { Badge, type BadgeTone } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { ScreenContainer } from '@/components/ui/screen-container';
import * as practicumService from '@/services/practicums';
import type { AssignedStudent, PracticumStatus } from '@/types/practicum';

const STATUS_TONE: Record<PracticumStatus, BadgeTone> = {
  active: 'success',
  completed: 'neutral',
  terminated: 'danger',
};

export default function SupervisorStudents() {
  const { data, isLoading } = useQuery({
    queryKey: ['students'],
    queryFn: practicumService.listMyStudents,
  });

  return (
    <ScreenContainer>
      <Text className="mb-1 text-2xl font-bold text-slate-900">Your Students</Text>
      <Text className="mb-6 text-sm text-slate-500">
        {data ? `${data.length} student${data.length === 1 ? '' : 's'} assigned to you` : ' '}
      </Text>

      {isLoading ? (
        <View className="items-center py-10">
          <ActivityIndicator color="#4f46e5" />
        </View>
      ) : !data || data.length === 0 ? (
        <EmptyState
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
