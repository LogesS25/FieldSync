import { Ionicons } from '@expo/vector-icons';
import { useQuery } from '@tanstack/react-query';
import { Pressable, Text, View } from 'react-native';

import { Card } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import { LoadingState } from '@/components/ui/loading-state';
import { PageHeader } from '@/components/ui/page-header';
import { ScreenContainer } from '@/components/ui/screen-container';
import { openAuthenticatedFile } from '@/lib/open-file';
import * as manualsService from '@/services/manuals';
import { useAuthStore } from '@/stores/auth-store';

export default function SupervisorResources() {
  const accessToken = useAuthStore((state) => state.accessToken);

  const { data: manual, isLoading, isError, refetch } = useQuery({
    queryKey: ['manuals', 'mine'],
    queryFn: manualsService.getMyManual,
  });

  const viewManual = async () => {
    if (!accessToken || !manual) return;
    try {
      await openAuthenticatedFile(manualsService.manualFileUrl(manual.id), accessToken, manual.filename);
    } catch {
      // Best-effort viewer — a failed open isn't worth a blocking error UI.
    }
  };

  return (
    <ScreenContainer scroll>
      <PageHeader
        icon="book-outline"
        title="Guidelines & Resources"
        description="The practicum guidance manual for your assigned university."
      />

      {isLoading ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState onRetry={() => refetch()} />
      ) : !manual ? (
        <EmptyState
          icon="book-outline"
          title="No manual available yet"
          description="The university hasn't uploaded a practicum guidance manual."
        />
      ) : (
        <Card>
          <View className="mb-3 flex-row items-center gap-3">
            <View className="h-11 w-11 items-center justify-center rounded-xl bg-brand-50">
              <Ionicons name="document-text-outline" size={20} color="#4f46e5" />
            </View>
            <View className="flex-1">
              <Text className="font-semibold text-slate-900">{manual.filename}</Text>
              <Text className="text-xs text-slate-500">
                Updated {new Date(manual.updatedAt).toLocaleDateString()}
              </Text>
            </View>
          </View>
          <Pressable
            onPress={viewManual}
            className="flex-row items-center justify-center gap-2 rounded-xl bg-brand-600 py-3.5 active:bg-brand-700"
          >
            <Ionicons name="eye-outline" size={17} color="#ffffff" />
            <Text className="font-semibold text-white">View Manual</Text>
          </Pressable>
        </Card>
      )}
    </ScreenContainer>
  );
}
