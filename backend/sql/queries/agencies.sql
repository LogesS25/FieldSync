-- name: CreateAgency :one
INSERT INTO agencies (name, institution_id) VALUES ($1, $2) RETURNING *;

-- name: ListAgencies :many
SELECT * FROM agencies ORDER BY name;

-- name: ListAgenciesForInstitution :many
SELECT * FROM agencies WHERE institution_id = $1 ORDER BY name;

-- name: GetAgencyByID :one
SELECT * FROM agencies WHERE id = $1;
