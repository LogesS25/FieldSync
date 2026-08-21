// Package attendance implements twice-daily (morning/evening) attendance
// recording with sequential agency-then-faculty approval (business
// requirements §9.1/§9.3). Hours are optional, free-form input — the
// attendance-to-hours calculation is explicitly TBD in the business
// requirements (§15) and must not be hard-coded; there is deliberately no
// "total hours" aggregate here, unlike the previous implementation.
package attendance

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/practicums"
)

var (
	ErrNotAssignedSupervisor = errors.New("you are not an assigned supervisor for this student")
	ErrAgencyReviewFirst     = errors.New("agency supervisor must approve before faculty review")
	ErrRecordNotFound        = errors.New("attendance record not found")
	ErrAlreadyReviewedByRole = errors.New("you have already reviewed this record")
)

type Service struct {
	queries    *sqlcgen.Queries
	practicums *practicums.Service
}

func NewService(queries *sqlcgen.Queries, practicumsService *practicums.Service) *Service {
	return &Service{queries: queries, practicums: practicumsService}
}

type CreateInput struct {
	StudentID      pgtype.UUID
	AttendanceDate pgtype.Date
	Session        sqlcgen.AttendanceSession
	Hours          *float64
}

func (s *Service) Create(ctx context.Context, in CreateInput) (sqlcgen.AttendanceRecord, error) {
	practicumID, err := s.practicums.GetActivePracticumID(ctx, in.StudentID)
	if err != nil {
		return sqlcgen.AttendanceRecord{}, err
	}

	var hours pgtype.Numeric
	if in.Hours != nil {
		hours, err = db.ParseHours(*in.Hours)
		if err != nil {
			return sqlcgen.AttendanceRecord{}, err
		}
	}

	return s.queries.CreateAttendanceRecord(ctx, sqlcgen.CreateAttendanceRecordParams{
		StudentID:      in.StudentID,
		PracticumID:    practicumID,
		AttendanceDate: in.AttendanceDate,
		Session:        in.Session,
		HoursLogged:    hours,
	})
}

func (s *Service) ListForStudent(ctx context.Context, studentID pgtype.UUID) ([]sqlcgen.AttendanceRecord, error) {
	return s.queries.ListAttendanceForStudent(ctx, studentID)
}

func (s *Service) ListPendingForSupervisor(ctx context.Context, supervisorID pgtype.UUID) ([]sqlcgen.AttendanceRecord, error) {
	return s.queries.ListPendingAttendanceForSupervisor(ctx, supervisorID)
}

func (s *Service) AgencyReview(ctx context.Context, recordID, supervisorID pgtype.UUID, approve bool) (sqlcgen.AttendanceRecord, error) {
	record, err := s.queries.GetAttendanceRecordByID(ctx, recordID)
	if err != nil {
		return sqlcgen.AttendanceRecord{}, ErrRecordNotFound
	}
	if err := s.requireAssignedSupervisor(ctx, record.PracticumID, supervisorID); err != nil {
		return sqlcgen.AttendanceRecord{}, err
	}
	if record.AgencyStatus != sqlcgen.ReviewDecisionPending {
		return sqlcgen.AttendanceRecord{}, ErrAlreadyReviewedByRole
	}

	decision := sqlcgen.ReviewDecisionRejected
	if approve {
		decision = sqlcgen.ReviewDecisionApproved
	}

	return s.queries.SetAttendanceAgencyDecision(ctx, sqlcgen.SetAttendanceAgencyDecisionParams{
		ID:               recordID,
		AgencyStatus:     decision,
		AgencyReviewedBy: supervisorID,
	})
}

func (s *Service) FacultyReview(ctx context.Context, recordID, supervisorID pgtype.UUID, approve bool) (sqlcgen.AttendanceRecord, error) {
	record, err := s.queries.GetAttendanceRecordByID(ctx, recordID)
	if err != nil {
		return sqlcgen.AttendanceRecord{}, ErrRecordNotFound
	}
	if err := s.requireAssignedSupervisor(ctx, record.PracticumID, supervisorID); err != nil {
		return sqlcgen.AttendanceRecord{}, err
	}
	if record.AgencyStatus != sqlcgen.ReviewDecisionApproved {
		return sqlcgen.AttendanceRecord{}, ErrAgencyReviewFirst
	}
	if record.FacultyStatus != sqlcgen.ReviewDecisionPending {
		return sqlcgen.AttendanceRecord{}, ErrAlreadyReviewedByRole
	}

	decision := sqlcgen.ReviewDecisionRejected
	if approve {
		decision = sqlcgen.ReviewDecisionApproved
	}

	return s.queries.SetAttendanceFacultyDecision(ctx, sqlcgen.SetAttendanceFacultyDecisionParams{
		ID:                recordID,
		FacultyStatus:     decision,
		FacultyReviewedBy: supervisorID,
	})
}

func (s *Service) requireAssignedSupervisor(ctx context.Context, practicumID, supervisorID pgtype.UUID) error {
	exists, err := s.queries.SupervisorAssignmentExists(ctx, sqlcgen.SupervisorAssignmentExistsParams{
		PracticumID:  practicumID,
		SupervisorID: supervisorID,
	})
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotAssignedSupervisor
	}
	return nil
}
