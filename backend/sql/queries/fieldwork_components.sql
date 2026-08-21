-- name: CreateFieldworkComponent :one
INSERT INTO fieldwork_components (institution_id, name) VALUES ($1, $2) RETURNING *;

-- name: ListFieldworkComponents :many
SELECT * FROM fieldwork_components ORDER BY name;

-- name: ListFieldworkComponentsForInstitution :many
SELECT * FROM fieldwork_components WHERE institution_id = $1 ORDER BY name;

-- name: GetFieldworkComponentByID :one
SELECT * FROM fieldwork_components WHERE id = $1;

-- name: UpdateFieldworkComponent :one
UPDATE fieldwork_components SET name = $2, updated_at = now() WHERE id = $1 RETURNING *;

-- name: DeleteFieldworkComponent :exec
DELETE FROM fieldwork_components WHERE id = $1;
