import { API_URL } from '@/lib/env';
import { ApiError, apiRequest } from '@/lib/api-client';
import type { Manual } from '@/types/manual';

export async function getMyManual(): Promise<Manual | null> {
  try {
    return await apiRequest<Manual>('/manuals/mine');
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) return null;
    throw err;
  }
}

export function manualFileUrl(id: string): string {
  return `${API_URL}/manuals/${id}/file`;
}
