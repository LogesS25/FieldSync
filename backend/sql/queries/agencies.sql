-- name: CreateAgency :one
INSERT INTO agencies (name) VALUES ($1) RETURNING *;

-- name: ListAgencies :many
SELECT * FROM agencies ORDER BY name;

-- name: GetAgencyByID :one
SELECT * FROM agencies WHERE id = $1;
