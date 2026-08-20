-- name: CreateAttendanceRecord :one
INSERT INTO attendance_records (student_id, practicum_id, attendance_date, hours_logged)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListAttendanceForStudent :many
SELECT * FROM attendance_records
WHERE student_id = $1
ORDER BY attendance_date DESC;

-- name: GetTotalHoursForStudent :one
-- Field-hours tracking (requirements goal #4) computed server-side — no
-- fetch-all-then-sum in Go.
SELECT COALESCE(SUM(hours_logged), 0)::numeric AS total_hours
FROM attendance_records
WHERE student_id = $1;
