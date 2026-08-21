-- name: CreateAgency :one
INSERT INTO agencies (name, institution_id) VALUES ($1, $2) RETURNING *;

-- name: ListAgencies :many
SELECT * FROM agencies ORDER BY name;

-- name: ListAgenciesForInstitution :many
SELECT * FROM agencies WHERE institution_id = $1 ORDER BY name;

-- name: GetAgencyByID :one
SELECT * FROM agencies WHERE id = $1;

-- name: UpdateAgency :one
UPDATE agencies SET name = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: DeleteAgency :exec
DELETE FROM agencies WHERE id = $1;
