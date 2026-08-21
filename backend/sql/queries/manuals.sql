-- name: UpsertManual :one
INSERT INTO manuals (institution_id, file_path, original_filename, uploaded_by)
VALUES ($1, $2, $3, $4)
ON CONFLICT (institution_id) DO UPDATE
SET file_path = EXCLUDED.file_path,
    original_filename = EXCLUDED.original_filename,
    uploaded_by = EXCLUDED.uploaded_by,
    updated_at = now()
RETURNING *;

-- name: GetManualByID :one
SELECT * FROM manuals WHERE id = $1;

-- name: GetManualForUser :one
-- Resolves the caller's own university (directly for students/faculty via
-- users.institution_id, indirectly for agency supervisors via their
-- agency's institution_id) and returns that university's current manual in
-- one round trip.
SELECT m.*
FROM manuals m
JOIN users u ON u.id = $1
LEFT JOIN agencies a ON a.id = u.agency_id
WHERE m.institution_id = COALESCE(u.institution_id, a.institution_id);

-- name: GetEffectiveInstitutionForUser :one
SELECT COALESCE(u.institution_id, a.institution_id)::uuid AS institution_id
FROM users u
LEFT JOIN agencies a ON a.id = u.agency_id
WHERE u.id = $1;

-- name: ListManuals :many
SELECT * FROM manuals ORDER BY institution_id;

-- name: DeleteManual :exec
DELETE FROM manuals WHERE institution_id = $1;
