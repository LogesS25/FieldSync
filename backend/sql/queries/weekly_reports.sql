-- name: CreateWeeklyReport :one
INSERT INTO weekly_reports (student_id, practicum_id, week_start_date, week_end_date, summary)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListWeeklyReportsForStudent :many
SELECT * FROM weekly_reports
WHERE student_id = $1
ORDER BY week_start_date DESC;
