import { API_URL } from '@/lib/env';
import { useAuthStore } from '@/stores/auth-store';

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

interface RequestOptions extends Omit<RequestInit, 'body'> {
  body?: unknown;
}

interface RefreshResponse {
  accessToken: string;
  refreshToken: string;
  user: import('@/types/auth').AuthUser;
}

function rawFetch(path: string, options: RequestOptions, accessToken: string | null) {
  return fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
      ...options.headers,
    },
    body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
  });
}

// Concurrent 401s share a single in-flight refresh instead of each firing
// their own /auth/refresh call (which would race to rotate the same token).
let refreshInFlight: Promise<string | null> | null = null;

function refreshAccessToken(): Promise<string | null> {
  const { refreshToken, setSession, clearSession } = useAuthStore.getState();
  if (!refreshToken) {
    return Promise.resolve(null);
  }

  if (!refreshInFlight) {
    refreshInFlight = (async () => {
      try {
        const response = await rawFetch('/auth/refresh', { method: 'POST', body: { refreshToken } }, null);
        if (!response.ok) {
          clearSession();
          return null;
        }
        const data = (await response.json()) as RefreshResponse;
        setSession(data.user, data.accessToken, data.refreshToken);
        return data.accessToken;
      } catch {
        clearSession();
        return null;
      } finally {
        refreshInFlight = null;
      }
    })();
  }

  return refreshInFlight;
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { accessToken } = useAuthStore.getState();
  let response = await rawFetch(path, options, accessToken);

  if (response.status === 401 && path !== '/auth/refresh' && path !== '/auth/login') {
    const newAccessToken = await refreshAccessToken();
    if (newAccessToken) {
      response = await rawFetch(path, options, newAccessToken);
    }
  }

  if (!response.ok) {
    const message = await response.text().catch(() => response.statusText);
    throw new ApiError(response.status, message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

function rawUpload(path: string, formData: FormData, accessToken: string | null) {
  return fetch(`${API_URL}${path}`, {
    method: 'POST',
    headers: {
      ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}),
    },
    body: formData,
  });
}

// Separate from apiRequest because multipart/form-data bodies must not be
// JSON-encoded and must not set Content-Type manually (fetch derives the
// boundary from the FormData itself).
export async function apiUpload<T>(path: string, formData: FormData): Promise<T> {
  const { accessToken } = useAuthStore.getState();
  let response = await rawUpload(path, formData, accessToken);

  if (response.status === 401) {
    const newAccessToken = await refreshAccessToken();
    if (newAccessToken) {
      response = await rawUpload(path, formData, newAccessToken);
    }
  }

  if (!response.ok) {
    const message = await response.text().catch(() => response.statusText);
    throw new ApiError(response.status, message);
  }

  return response.json() as Promise<T>;
}
