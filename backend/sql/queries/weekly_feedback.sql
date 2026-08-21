-- name: CreateWeeklyFeedback :one
INSERT INTO weekly_feedback (practicum_id, supervisor_id, week_start_date, feedback)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListFeedbackForStudent :many
-- Joins through practicums so the student never needs to know their own
-- practicum_id to see feedback about them.
SELECT wf.*
FROM weekly_feedback wf
JOIN practicums p ON p.id = wf.practicum_id
WHERE p.student_id = $1
ORDER BY wf.week_start_date DESC, wf.created_at DESC;

-- name: ListFeedbackFromSupervisor :many
SELECT * FROM weekly_feedback WHERE supervisor_id = $1 ORDER BY week_start_date DESC;
