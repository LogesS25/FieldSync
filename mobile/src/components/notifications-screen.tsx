import { Ionicons } from '@expo/vector-icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Pressable, Text, View } from 'react-native';

import { Card } from '@/components/ui/card';
import { EmptyState } from '@/components/ui/empty-state';
import { ErrorState } from '@/components/ui/error-state';
import { LoadingState } from '@/components/ui/loading-state';
import { PageHeader } from '@/components/ui/page-header';
import { ScreenContainer } from '@/components/ui/screen-container';
import * as notificationsService from '@/services/notifications';
import type { AppNotification } from '@/types/notification';

// Shared by both role route files — the notification list, mark-read, and
// mark-all-read behavior has no role-specific variation, so this avoids
// maintaining two identical copies.
export function NotificationsScreen() {
  const queryClient = useQueryClient();

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['notifications'],
    queryFn: notificationsService.listNotifications,
  });

  const markReadMutation = useMutation({
    mutationFn: notificationsService.markNotificationRead,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  });

  const markAllReadMutation = useMutation({
    mutationFn: notificationsService.markAllNotificationsRead,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['notifications'] }),
  });

  const unreadCount = data?.filter((n) => !n.readAt).length ?? 0;

  return (
    <ScreenContainer scroll>
      <PageHeader
        icon="notifications-outline"
        title="Notifications"
        description="Updates on your requests, reviews, and feedback."
        action={
          unreadCount > 0 ? (
            <Pressable
              onPress={() => markAllReadMutation.mutate()}
              disabled={markAllReadMutation.isPending}
              className="rounded-full border border-slate-200 bg-white px-3 py-1.5 active:bg-slate-50"
            >
              <Text className="text-xs font-semibold text-brand-600">Mark all read</Text>
            </Pressable>
          ) : undefined
        }
      />

      {isLoading ? (
        <LoadingState />
      ) : isError ? (
        <ErrorState onRetry={() => refetch()} />
      ) : !data || data.length === 0 ? (
        <EmptyState
          icon="notifications-outline"
          title="No notifications yet"
          description="Updates on your requests, reviews, and feedback will show up here."
        />
      ) : (
        <View className="gap-2.5">
          {data.map((notification) => (
            <NotificationCard
              key={notification.id}
              notification={notification}
              onPress={() => {
                if (!notification.readAt) markReadMutation.mutate(notification.id);
              }}
            />
          ))}
        </View>
      )}
    </ScreenContainer>
  );
}

function NotificationCard({
  notification,
  onPress,
}: {
  notification: AppNotification;
  onPress: () => void;
}) {
  const isUnread = !notification.readAt;

  return (
    <Pressable onPress={onPress} disabled={!isUnread}>
      <Card className={isUnread ? 'border-brand-200 bg-brand-50/40' : undefined}>
        <View className="flex-row items-start gap-3">
          <View
            className={`mt-0.5 h-8 w-8 items-center justify-center rounded-full ${
              isUnread ? 'bg-brand-100' : 'bg-slate-100'
            }`}
          >
            <Ionicons
              name={isUnread ? 'notifications' : 'notifications-outline'}
              size={15}
              color={isUnread ? '#4338ca' : '#94a3b8'}
            />
          </View>
          <View className="flex-1">
            <Text className={`text-sm ${isUnread ? 'font-semibold text-slate-900' : 'text-slate-600'}`}>
              {notification.message}
            </Text>
            <Text className="mt-1 text-xs text-slate-400">
              {new Date(notification.createdAt).toLocaleString()}
            </Text>
          </View>
          {isUnread ? <View className="mt-1.5 h-2 w-2 rounded-full bg-brand-600" /> : null}
        </View>
      </Card>
    </Pressable>
  );
}
