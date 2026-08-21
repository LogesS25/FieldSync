import { apiRequest } from '@/lib/api-client';
import type { WeeklyFeedback } from '@/types/feedback';

export function listFeedbackForMe(): Promise<WeeklyFeedback[]> {
  return apiRequest<WeeklyFeedback[]>('/feedback');
}

export function listFeedbackIGave(): Promise<WeeklyFeedback[]> {
  return apiRequest<WeeklyFeedback[]>('/feedback/mine');
}

export function submitFeedback(input: {
  practicumId: string;
  weekStartDate: string;
  feedback: string;
}): Promise<WeeklyFeedback> {
  return apiRequest<WeeklyFeedback>('/feedback', { method: 'POST', body: input });
}
