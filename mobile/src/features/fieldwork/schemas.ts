import { z } from 'zod';

const isoDate = z
  .string()
  .min(1, 'Date is required')
  .regex(/^\d{4}-\d{2}-\d{2}$/, 'Use YYYY-MM-DD');

export const fieldActivitySchema = z.object({
  activityDate: isoDate,
  description: z.string().min(1, 'Description is required'),
});
export type FieldActivityFormValues = z.infer<typeof fieldActivitySchema>;

export const attendanceSchema = z.object({
  attendanceDate: isoDate,
  session: z.enum(['morning', 'evening']),
  hours: z
    .string()
    .optional()
    .refine((v) => !v || !Number.isNaN(Number(v)), 'Must be a number')
    .refine((v) => !v || (Number(v) > 0 && Number(v) <= 24), 'Must be between 0 and 24'),
});
export type AttendanceFormValues = z.infer<typeof attendanceSchema>;

export const consolidatedReportSchema = z.object({
  summary: z.string().min(1, 'Summary is required'),
});
export type ConsolidatedReportFormValues = z.infer<typeof consolidatedReportSchema>;

export const teamRequestSchema = z.object({
  agencyId: z.string().min(1, 'Select an agency'),
  facultySupervisorId: z.string().min(1, 'Select a faculty supervisor'),
  agencySupervisorId: z.string().min(1, 'Select an agency supervisor'),
  fieldworkDescription: z.string().min(1, 'Describe the fieldwork'),
  startDate: isoDate,
});
export type TeamRequestFormValues = z.infer<typeof teamRequestSchema>;
