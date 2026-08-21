DROP TABLE daily_reports;

CREATE TYPE verification_status AS ENUM ('pending', 'verified', 'rejected');

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
