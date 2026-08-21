-- Reworks Phase 3/4 to match the updated business requirements document
-- (social_work_field_practicum_business_requirements.md, 2026-08-20). See
-- docs/ARCHITECTURE.md §3a for the full rationale behind each change below.

-- 1. Agencies become university-scoped (a university owns its own agency
-- list — business doc §5). Backfill existing rows onto the first
-- institution before making the column required, since this is pre-launch
-- dev data, not a real migration concern.
ALTER TABLE agencies ADD COLUMN institution_id UUID REFERENCES institutions(id);
UPDATE agencies SET institution_id = (SELECT id FROM institutions ORDER BY created_at LIMIT 1)
WHERE institution_id IS NULL AND EXISTS (SELECT 1 FROM institutions);
ALTER TABLE agencies ALTER COLUMN institution_id SET NOT NULL;
CREATE INDEX idx_agencies_institution_id ON agencies (institution_id);

-- 2. Practicum team formation is student-initiated with independent
-- accept/reject per supervisor (business doc §8), not admin-unilateral.
-- "Fieldwork" selection structure is explicitly TBD in the business doc
-- (§7) — stored as free text rather than a rigid catalog until the
-- university stakeholder defines it.
CREATE TYPE team_request_decision AS ENUM ('pending', 'accepted', 'rejected');

CREATE TABLE practicum_team_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    institution_id UUID NOT NULL REFERENCES institutions(id),
    agency_id UUID NOT NULL REFERENCES agencies(id),
    faculty_supervisor_id UUID NOT NULL REFERENCES users(id),
    agency_supervisor_id UUID NOT NULL REFERENCES users(id),
    fieldwork_description TEXT NOT NULL,
    start_date DATE NOT NULL,
    faculty_decision team_request_decision NOT NULL DEFAULT 'pending',
    agency_decision team_request_decision NOT NULL DEFAULT 'pending',
    -- Set once both decisions are 'accepted' and the Practicum/Placement/
    -- SupervisorAssignment rows have been created from this request.
    formed_practicum_id UUID REFERENCES practicums(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_team_requests_student_id ON practicum_team_requests (student_id);
CREATE INDEX idx_team_requests_faculty_supervisor_id ON practicum_team_requests (faculty_supervisor_id);
CREATE INDEX idx_team_requests_agency_supervisor_id ON practicum_team_requests (agency_supervisor_id);

-- 3. Attendance becomes twice-daily (morning/evening) with sequential
-- agency-then-faculty approval (business doc §9.1/§9.3), replacing the
-- single-record-per-day/single-status model. `hours_logged` becomes
-- optional free-form input, not an authoritative total — the
-- attendance-to-hours calculation is explicitly TBD (business doc §15) and
-- must not be hard-coded.
CREATE TYPE attendance_session AS ENUM ('morning', 'evening');
CREATE TYPE review_decision AS ENUM ('pending', 'approved', 'rejected');

ALTER TABLE attendance_records DROP CONSTRAINT attendance_records_student_id_attendance_date_key;
ALTER TABLE attendance_records ADD COLUMN session attendance_session NOT NULL DEFAULT 'morning';
ALTER TABLE attendance_records ALTER COLUMN session DROP DEFAULT;
ALTER TABLE attendance_records ADD CONSTRAINT attendance_records_student_date_session_key
    UNIQUE (student_id, attendance_date, session);

ALTER TABLE attendance_records ALTER COLUMN hours_logged DROP NOT NULL;
ALTER TABLE attendance_records DROP CONSTRAINT attendance_records_hours_logged_check;
ALTER TABLE attendance_records ADD CONSTRAINT attendance_records_hours_logged_check
    CHECK (hours_logged IS NULL OR (hours_logged > 0 AND hours_logged <= 24));

ALTER TABLE attendance_records DROP COLUMN verification_status;
ALTER TABLE attendance_records DROP COLUMN verified_by;
ALTER TABLE attendance_records DROP COLUMN verified_at;

ALTER TABLE attendance_records ADD COLUMN agency_status review_decision NOT NULL DEFAULT 'pending';
ALTER TABLE attendance_records ADD COLUMN agency_reviewed_by UUID REFERENCES users(id);
ALTER TABLE attendance_records ADD COLUMN agency_reviewed_at TIMESTAMPTZ;
ALTER TABLE attendance_records ADD COLUMN faculty_status review_decision NOT NULL DEFAULT 'pending';
ALTER TABLE attendance_records ADD COLUMN faculty_reviewed_by UUID REFERENCES users(id);
ALTER TABLE attendance_records ADD COLUMN faculty_reviewed_at TIMESTAMPTZ;

-- 4. Weekly reports are replaced by a single consolidated report per
-- practicum (business doc §13), reviewed agency-then-faculty like
-- attendance. Nothing depends on weekly_reports pre-launch, so it's
-- dropped rather than migrated.
DROP TABLE weekly_reports;
DROP TYPE report_status;

CREATE TABLE consolidated_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    practicum_id UUID NOT NULL UNIQUE REFERENCES practicums(id) ON DELETE CASCADE,
    summary TEXT NOT NULL,
    agency_status review_decision NOT NULL DEFAULT 'pending',
    agency_reviewed_by UUID REFERENCES users(id),
    agency_reviewed_at TIMESTAMPTZ,
    faculty_status review_decision NOT NULL DEFAULT 'pending',
    faculty_reviewed_by UUID REFERENCES users(id),
    faculty_reviewed_at TIMESTAMPTZ,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
