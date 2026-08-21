-- name: CreateConsolidatedReport :one
INSERT INTO consolidated_reports (student_id, practicum_id, summary)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetConsolidatedReportForStudent :one
SELECT * FROM consolidated_reports WHERE student_id = $1;

-- name: GetConsolidatedReportByID :one
SELECT * FROM consolidated_reports WHERE id = $1;

-- name: SetConsolidatedReportAgencyDecision :one
UPDATE consolidated_reports
SET agency_status = $2, agency_reviewed_by = $3, agency_reviewed_at = now(), updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetConsolidatedReportFacultyDecision :one
UPDATE consolidated_reports
SET faculty_status = $2, faculty_reviewed_by = $3, faculty_reviewed_at = now(), updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListPendingConsolidatedReportsForSupervisor :many
SELECT cr.*
FROM consolidated_reports cr
JOIN practicums p ON p.id = cr.practicum_id
JOIN supervisor_assignments sa ON sa.practicum_id = p.id
JOIN users u ON u.id = sa.supervisor_id
WHERE sa.supervisor_id = $1
  AND (
    (u.role = 'agency_supervisor' AND cr.agency_status = 'pending')
    OR (u.role = 'faculty_supervisor' AND cr.agency_status = 'approved' AND cr.faculty_status = 'pending')
  )
ORDER BY cr.submitted_at DESC;
