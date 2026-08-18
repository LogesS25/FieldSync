-- Phase 1 foundation migration.
-- Only the users table is created now, to prove the migration/DB pipeline
-- works end-to-end. Full domain schema (institutions, agencies, practicums,
-- field activities, competencies, etc.) is designed in Phase 2/3 per the
-- roadmap in docs/ARCHITECTURE.md.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM (
    'student',
    'faculty_supervisor',
    'agency_supervisor',
    'administrator'
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role user_role NOT NULL,
    full_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
