-- name: CreatePracticum :one
INSERT INTO practicums (student_id, institution_id, start_date, end_date)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreatePlacement :one
INSERT INTO placements (practicum_id, agency_id, start_date, end_date)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateSupervisorAssignment :one
INSERT INTO supervisor_assignments (practicum_id, supervisor_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetPracticumByID :one
SELECT * FROM practicums WHERE id = $1;

-- name: SupervisorAssignmentExists :one
SELECT EXISTS (
    SELECT 1 FROM supervisor_assignments
    WHERE practicum_id = $1 AND supervisor_id = $2
);

-- name: GetPracticumSummaryForStudent :one
-- Single round trip: current placement resolved via LATERAL (most recent
-- placement row), assigned supervisors aggregated server-side with
-- json_agg — no application-level looping or joining needed to build the
-- response.
SELECT
    p.id AS practicum_id,
    p.status,
    p.start_date,
    p.end_date,
    i.name AS institution_name,
    cp.agency_id,
    ag.name AS agency_name,
    cp.start_date AS placement_start_date,
    cp.end_date AS placement_end_date,
    COALESCE(sup.supervisors, '[]'::json) AS supervisors
FROM practicums p
JOIN institutions i ON i.id = p.institution_id
LEFT JOIN LATERAL (
    SELECT agency_id, start_date, end_date
    FROM placements
    WHERE practicum_id = p.id
    ORDER BY start_date DESC
    LIMIT 1
) cp ON true
LEFT JOIN agencies ag ON ag.id = cp.agency_id
LEFT JOIN LATERAL (
    SELECT json_agg(json_build_object(
        'id', u.id,
        'fullName', u.full_name,
        'role', u.role
    )) AS supervisors
    FROM supervisor_assignments sa
    JOIN users u ON u.id = sa.supervisor_id
    WHERE sa.practicum_id = p.id
) sup ON true
WHERE p.student_id = $1 AND p.status = 'active'
ORDER BY p.start_date DESC
LIMIT 1;

-- name: ListStudentsForSupervisor :many
-- One query per request, not one query per student: current placement
-- resolved via LATERAL join instead of an N+1 fetch-then-loop.
SELECT
    u.id AS student_id,
    u.full_name AS student_name,
    u.email AS student_email,
    p.id AS practicum_id,
    p.status AS practicum_status,
    p.start_date AS practicum_start_date,
    p.end_date AS practicum_end_date,
    i.name AS institution_name,
    ag.id AS agency_id,
    ag.name AS agency_name
FROM supervisor_assignments sa
JOIN practicums p ON p.id = sa.practicum_id
JOIN users u ON u.id = p.student_id
JOIN institutions i ON i.id = p.institution_id
LEFT JOIN LATERAL (
    SELECT agency_id
    FROM placements
    WHERE practicum_id = p.id
    ORDER BY start_date DESC
    LIMIT 1
) cp ON true
LEFT JOIN agencies ag ON ag.id = cp.agency_id
WHERE sa.supervisor_id = $1
ORDER BY u.full_name;
