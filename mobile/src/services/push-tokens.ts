import { apiRequest } from '@/lib/api-client';

export function registerPushToken(token: string): Promise<void> {
  return apiRequest<void>('/push-tokens', { method: 'POST', body: { token } });
}

export function unregisterPushToken(token: string): Promise<void> {
  return apiRequest<void>('/push-tokens', { method: 'DELETE', body: { token } });
}
