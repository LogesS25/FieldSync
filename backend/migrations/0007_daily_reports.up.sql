-- Replaces field_activities (free-text daily log) with daily_reports (the
-- business requirements' actual "Daily Handwritten Report" — a PDF upload
-- of a handwritten page, §10), with the same agency-then-faculty sequential
-- review pattern already used by attendance_records/consolidated_reports.
-- Dev-only environment, no real field_activities data to migrate (see
-- AGENTS.md precedent for migration 0005/0006).

DROP TABLE field_activities;
DROP TYPE verification_status;

CREATE TABLE daily_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    practicum_id UUID NOT NULL REFERENCES practicums(id) ON DELETE CASCADE,
    report_date DATE NOT NULL,
    file_path TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    agency_status review_decision NOT NULL DEFAULT 'pending',
    agency_reviewed_by UUID REFERENCES users(id),
    agency_reviewed_at TIMESTAMPTZ,
    faculty_status review_decision NOT NULL DEFAULT 'pending',
    faculty_reviewed_by UUID REFERENCES users(id),
    faculty_reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (student_id, report_date)
);

CREATE INDEX idx_daily_reports_practicum_id ON daily_reports (practicum_id);
