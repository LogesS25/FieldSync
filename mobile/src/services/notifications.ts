import { apiRequest } from '@/lib/api-client';
import type { AppNotification } from '@/types/notification';

export function listNotifications(): Promise<AppNotification[]> {
  return apiRequest<AppNotification[]>('/notifications');
}

export function markNotificationRead(id: string): Promise<AppNotification> {
  return apiRequest<AppNotification>(`/notifications/${id}/read`, { method: 'POST' });
}

export function markAllNotificationsRead(): Promise<void> {
  return apiRequest<void>('/notifications/read-all', { method: 'POST' });
}
