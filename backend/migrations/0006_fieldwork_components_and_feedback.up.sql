-- Closes three gaps identified after the 2026-08-20 business requirements
-- update (see docs/ARCHITECTURE.md §3a and the follow-up entry below it):
-- fieldwork-component selection, mandatory weekly feedback, and
-- consolidated-report resubmission-after-rejection.

-- 1. Fieldwork components: university-defined list students select from
-- when forming a practicum team (business doc §7). University-scoped like
-- agencies.
CREATE TABLE fieldwork_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID NOT NULL REFERENCES institutions(id),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (institution_id, name)
);

CREATE INDEX idx_fieldwork_components_institution_id ON fieldwork_components (institution_id);

-- practicum_team_requests had zero rows in every environment this has run
-- in (dev-only, pre-launch), so this is added NOT NULL directly rather than
-- backfilled.
ALTER TABLE practicum_team_requests
    ADD COLUMN fieldwork_component_id UUID NOT NULL REFERENCES fieldwork_components(id);

-- 2. Weekly feedback: both supervisors must provide feedback every weekend
-- (business doc §12). One row per (practicum, supervisor, week) — mirrors
-- the old weekly_reports shape, but for feedback specifically, and
-- supervisor-authored rather than student-authored.
CREATE TABLE weekly_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practicum_id UUID NOT NULL REFERENCES practicums(id) ON DELETE CASCADE,
    supervisor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    week_start_date DATE NOT NULL,
    feedback TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (practicum_id, supervisor_id, week_start_date)
);

CREATE INDEX idx_weekly_feedback_practicum_id ON weekly_feedback (practicum_id);
