import { apiRequest } from '@/lib/api-client';

interface HealthResponse {
  status: string;
  database: string;
}

export function getHealth(): Promise<HealthResponse> {
  return apiRequest<HealthResponse>('/health');
}
