-- name: CreateAttendanceRecord :one
INSERT INTO attendance_records (student_id, practicum_id, attendance_date, session, hours_logged)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListAttendanceForStudent :many
SELECT * FROM attendance_records
WHERE student_id = $1
ORDER BY attendance_date DESC, session;

-- name: GetAttendanceRecordByID :one
SELECT * FROM attendance_records WHERE id = $1;

-- name: SetAttendanceAgencyDecision :one
UPDATE attendance_records
SET agency_status = $2, agency_reviewed_by = $3, agency_reviewed_at = now(), updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetAttendanceFacultyDecision :one
UPDATE attendance_records
SET faculty_status = $2, faculty_reviewed_by = $3, faculty_reviewed_at = now(), updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListPendingAttendanceForSupervisor :many
-- Attendance awaiting this supervisor's review, across all their assigned
-- students — one query joining supervisor_assignments -> practicums ->
-- attendance_records rather than a fetch-then-filter per student.
SELECT ar.*
FROM attendance_records ar
JOIN practicums p ON p.id = ar.practicum_id
JOIN supervisor_assignments sa ON sa.practicum_id = p.id
JOIN users u ON u.id = sa.supervisor_id
WHERE sa.supervisor_id = $1
  AND (
    (u.role = 'agency_supervisor' AND ar.agency_status = 'pending')
    OR (u.role = 'faculty_supervisor' AND ar.agency_status = 'approved' AND ar.faculty_status = 'pending')
  )
ORDER BY ar.attendance_date DESC, ar.session;
