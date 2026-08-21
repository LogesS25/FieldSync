-- name: CreateDailyReport :one
INSERT INTO daily_reports (student_id, practicum_id, report_date, file_path, original_filename)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListDailyReportsForStudent :many
SELECT * FROM daily_reports
WHERE student_id = $1
ORDER BY report_date DESC;

-- name: GetDailyReportByID :one
SELECT * FROM daily_reports WHERE id = $1;

-- name: SetDailyReportAgencyDecision :one
UPDATE daily_reports
SET agency_status = $2, agency_reviewed_by = $3, agency_reviewed_at = now(), updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetDailyReportFacultyDecision :one
UPDATE daily_reports
SET faculty_status = $2, faculty_reviewed_by = $3, faculty_reviewed_at = now(), updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListPendingDailyReportsForSupervisor :many
-- Daily reports awaiting this supervisor's review, across all their
-- assigned students — one query joining supervisor_assignments ->
-- practicums -> daily_reports rather than a fetch-then-filter per student.
SELECT dr.*
FROM daily_reports dr
JOIN practicums p ON p.id = dr.practicum_id
JOIN supervisor_assignments sa ON sa.practicum_id = p.id
JOIN users u ON u.id = sa.supervisor_id
WHERE sa.supervisor_id = $1
  AND (
    (u.role = 'agency_supervisor' AND dr.agency_status = 'pending')
    OR (u.role = 'faculty_supervisor' AND dr.agency_status = 'approved' AND dr.faculty_status = 'pending')
  )
ORDER BY dr.report_date DESC;
