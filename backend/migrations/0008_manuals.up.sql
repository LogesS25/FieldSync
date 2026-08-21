-- Guidelines & Manuals (business requirements §17). Exact manual
-- versioning/archiving/visibility rules are explicitly TBD, so this keeps
-- to what IS specified: one current manual per university, replaced on
-- re-upload (UNIQUE(institution_id) + upsert), visible to every stakeholder
-- of that university (student, faculty supervisor, and agency supervisors
-- at agencies belonging to it).

CREATE TABLE manuals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    institution_id UUID NOT NULL UNIQUE REFERENCES institutions(id),
    file_path TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
