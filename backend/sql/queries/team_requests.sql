-- name: CreateTeamRequest :one
INSERT INTO practicum_team_requests
    (student_id, institution_id, agency_id, faculty_supervisor_id, agency_supervisor_id, fieldwork_description, start_date)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetTeamRequestByID :one
SELECT * FROM practicum_team_requests WHERE id = $1;

-- name: ListTeamRequestsForStudent :many
SELECT * FROM practicum_team_requests WHERE student_id = $1 ORDER BY created_at DESC;

-- name: ListTeamRequestsForSupervisor :many
-- A supervisor sees requests naming them as either the faculty or agency
-- supervisor, regardless of which role slot — one query instead of two.
SELECT * FROM practicum_team_requests
WHERE faculty_supervisor_id = $1 OR agency_supervisor_id = $1
ORDER BY created_at DESC;

-- name: SetFacultyDecision :one
UPDATE practicum_team_requests
SET faculty_decision = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetAgencyDecision :one
UPDATE practicum_team_requests
SET agency_decision = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: MarkTeamRequestFormed :exec
UPDATE practicum_team_requests SET formed_practicum_id = $2, updated_at = now() WHERE id = $1;
