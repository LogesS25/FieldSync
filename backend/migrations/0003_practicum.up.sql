-- Institutions, Agencies, and the relationships that make "assigned
-- students" a real queryable fact: Practicum (student's enrollment period at
-- an institution), Placement (which agency that practicum is at), and
-- SupervisorAssignment (which supervisors watch that practicum). See
-- docs/ARCHITECTURE.md §3 for the full domain model rationale.

CREATE TYPE practicum_status AS ENUM ('active', 'completed', 'terminated');

CREATE TABLE institutions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE agencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- institution_id: the student's or faculty supervisor's home institution.
-- agency_id: the agency supervisor's employer. Nullable because a freshly
-- registered user has neither until an administrator assigns one.
ALTER TABLE users
    ADD COLUMN institution_id UUID REFERENCES institutions(id),
    ADD COLUMN agency_id UUID REFERENCES agencies(id);

CREATE INDEX idx_users_institution_id ON users (institution_id) WHERE institution_id IS NOT NULL;
CREATE INDEX idx_users_agency_id ON users (agency_id) WHERE agency_id IS NOT NULL;

CREATE TABLE practicums (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    institution_id UUID NOT NULL REFERENCES institutions(id),
    status practicum_status NOT NULL DEFAULT 'active',
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_practicums_student_id ON practicums (student_id);

CREATE TABLE placements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practicum_id UUID NOT NULL REFERENCES practicums(id) ON DELETE CASCADE,
    agency_id UUID NOT NULL REFERENCES agencies(id),
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Supports "most recent placement for this practicum" (current agency)
-- lookups via LATERAL join without a sequential scan.
CREATE INDEX idx_placements_practicum_id_start_date ON placements (practicum_id, start_date DESC);

CREATE TABLE supervisor_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    practicum_id UUID NOT NULL REFERENCES practicums(id) ON DELETE CASCADE,
    supervisor_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (practicum_id, supervisor_id)
);

CREATE INDEX idx_supervisor_assignments_supervisor_id ON supervisor_assignments (supervisor_id);
