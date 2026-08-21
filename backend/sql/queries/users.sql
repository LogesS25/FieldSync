-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, role, full_name, institution_id, agency_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListFacultySupervisorsForInstitution :many
SELECT * FROM users
WHERE role = 'faculty_supervisor' AND institution_id = $1
ORDER BY full_name;

-- name: ListAgencySupervisorsForAgency :many
SELECT * FROM users
WHERE role = 'agency_supervisor' AND agency_id = $1
ORDER BY full_name;
