DROP TABLE IF EXISTS consolidated_reports;

CREATE TYPE report_status AS ENUM ('submitted', 'reviewed');
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

ALTER TABLE attendance_records DROP COLUMN faculty_reviewed_at;
ALTER TABLE attendance_records DROP COLUMN faculty_reviewed_by;
ALTER TABLE attendance_records DROP COLUMN faculty_status;
ALTER TABLE attendance_records DROP COLUMN agency_reviewed_at;
ALTER TABLE attendance_records DROP COLUMN agency_reviewed_by;
ALTER TABLE attendance_records DROP COLUMN agency_status;

ALTER TABLE attendance_records ADD COLUMN verification_status verification_status NOT NULL DEFAULT 'pending';
ALTER TABLE attendance_records ADD COLUMN verified_by UUID REFERENCES users(id);
ALTER TABLE attendance_records ADD COLUMN verified_at TIMESTAMPTZ;

ALTER TABLE attendance_records DROP CONSTRAINT attendance_records_hours_logged_check;
ALTER TABLE attendance_records ADD CONSTRAINT attendance_records_hours_logged_check
    CHECK (hours_logged > 0 AND hours_logged <= 24);
ALTER TABLE attendance_records ALTER COLUMN hours_logged SET NOT NULL;

ALTER TABLE attendance_records DROP CONSTRAINT attendance_records_student_date_session_key;
ALTER TABLE attendance_records DROP COLUMN session;
ALTER TABLE attendance_records ADD CONSTRAINT attendance_records_student_id_attendance_date_key
    UNIQUE (student_id, attendance_date);

DROP TYPE review_decision;
DROP TYPE attendance_session;

DROP TABLE IF EXISTS practicum_team_requests;
DROP TYPE IF EXISTS team_request_decision;

ALTER TABLE agencies DROP COLUMN institution_id;
