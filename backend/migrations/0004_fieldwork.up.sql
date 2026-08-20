-- Student Fieldwork (Phase 4): daily field activity logs, attendance with
-- hours (feeds the "field hours" tracking goal), and weekly report
-- submissions. Verification of activities/attendance by supervisors is
-- Phase 5 — verification_status exists now so Phase 5 doesn't need a
-- migration just to add it, but nothing sets it to anything but 'pending'
-- yet.

CREATE TYPE verification_status AS ENUM ('pending', 'verified', 'rejected');
CREATE TYPE report_status AS ENUM ('submitted', 'reviewed');

CREATE TABLE field_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    practicum_id UUID NOT NULL REFERENCES practicums(id) ON DELETE CASCADE,
    activity_date DATE NOT NULL,
    description TEXT NOT NULL,
    verification_status verification_status NOT NULL DEFAULT 'pending',
    verified_by UUID REFERENCES users(id),
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_field_activities_student_id_date ON field_activities (student_id, activity_date DESC);
CREATE INDEX idx_field_activities_practicum_id ON field_activities (practicum_id);

CREATE TABLE attendance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    practicum_id UUID NOT NULL REFERENCES practicums(id) ON DELETE CASCADE,
    attendance_date DATE NOT NULL,
    hours_logged NUMERIC(4, 2) NOT NULL CHECK (hours_logged > 0 AND hours_logged <= 24),
    verification_status verification_status NOT NULL DEFAULT 'pending',
    verified_by UUID REFERENCES users(id),
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (student_id, attendance_date)
);

CREATE INDEX idx_attendance_records_practicum_id ON attendance_records (practicum_id);

CREATE TABLE weekly_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    practicum_id UUID NOT NULL REFERENCES practicums(id) ON DELETE CASCADE,
    week_start_date DATE NOT NULL,
    week_end_date DATE NOT NULL,
    summary TEXT NOT NULL,
    status report_status NOT NULL DEFAULT 'submitted',
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (student_id, week_start_date)
);

CREATE INDEX idx_weekly_reports_student_id_week ON weekly_reports (student_id, week_start_date DESC);
