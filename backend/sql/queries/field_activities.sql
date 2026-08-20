-- name: CreateFieldActivity :one
INSERT INTO field_activities (student_id, practicum_id, activity_date, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListFieldActivitiesForStudent :many
SELECT * FROM field_activities
WHERE student_id = $1
ORDER BY activity_date DESC, created_at DESC;
