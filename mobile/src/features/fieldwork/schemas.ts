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

// `hours` stays a string at the form-state layer (TextInput only ever
// produces strings) and is parsed to a number at submit time — using
// z.coerce here would type the RHF field as `number` while the actual
// runtime value during editing is a string, which trips up
// @hookform/resolvers' generic inference.
export const attendanceSchema = z.object({
  attendanceDate: isoDate,
  hours: z
    .string()
    .min(1, 'Hours is required')
    .refine((v) => !Number.isNaN(Number(v)), 'Must be a number')
    .refine((v) => Number(v) > 0 && Number(v) <= 24, 'Must be between 0 and 24'),
});
export type AttendanceFormValues = z.infer<typeof attendanceSchema>;

export const weeklyReportSchema = z.object({
  weekStartDate: isoDate,
  weekEndDate: isoDate,
  summary: z.string().min(1, 'Summary is required'),
});
export type WeeklyReportFormValues = z.infer<typeof weeklyReportSchema>;
