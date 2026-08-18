-- name: CreateInstitution :one
INSERT INTO institutions (name) VALUES ($1) RETURNING *;

-- name: ListInstitutions :many
SELECT * FROM institutions ORDER BY name;

-- name: GetInstitutionByID :one
SELECT * FROM institutions WHERE id = $1;
