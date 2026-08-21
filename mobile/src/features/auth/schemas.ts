import { z } from 'zod';

export const loginSchema = z.object({
  email: z.string().min(1, 'Email is required').email('Enter a valid email'),
  password: z.string().min(1, 'Password is required'),
});

export type LoginFormValues = z.infer<typeof loginSchema>;

// Administrator is intentionally excluded — accounts for that role are
// provisioned separately, not self-registered (see docs/ARCHITECTURE.md §8).
// Every stakeholder is tied to a university, directly (student, faculty) or
// via their agency (agency supervisor) — business requirements §3 — so
// institutionId/agencyId are required per role, validated below.
export const registerSchema = z
  .object({
    fullName: z.string().min(1, 'Full name is required'),
    email: z.string().min(1, 'Email is required').email('Enter a valid email'),
    password: z.string().min(8, 'Password must be at least 8 characters'),
    role: z.enum(['student', 'faculty_supervisor', 'agency_supervisor']),
    institutionId: z.string().optional(),
    agencyId: z.string().optional(),
  })
  .refine((data) => data.role === 'agency_supervisor' || !!data.institutionId, {
    message: 'Select your university',
    path: ['institutionId'],
  })
  .refine((data) => data.role !== 'agency_supervisor' || !!data.agencyId, {
    message: 'Select your agency',
    path: ['agencyId'],
  });

export type RegisterFormValues = z.infer<typeof registerSchema>;
