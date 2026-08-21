import type { ReviewDecision } from '@/types/fieldwork';

export interface DailyReport {
  id: string;
  practicumId: string;
  reportDate: string;
  filename: string;
  agencyStatus: ReviewDecision;
  facultyStatus: ReviewDecision;
}
