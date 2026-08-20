package practicums

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

var (
	ErrStudentNotFound     = errors.New("student not found")
	ErrUserIsNotStudent    = errors.New("target user is not a student")
	ErrPracticumNotFound   = errors.New("practicum not found")
	ErrSupervisorNotFound  = errors.New("supervisor not found")
	ErrUserIsNotSupervisor = errors.New("target user is not a faculty or agency supervisor")
	ErrAlreadyAssigned     = errors.New("supervisor is already assigned to this practicum")
	ErrNoActivePracticum   = errors.New("student has no active practicum")
)

type Service struct {
	queries *sqlcgen.Queries
}

func NewService(queries *sqlcgen.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) CreatePracticum(ctx context.Context, studentID, institutionID pgtype.UUID, startDate, endDate pgtype.Date) (sqlcgen.Practicum, error) {
	student, err := s.queries.GetUserByID(ctx, studentID)
	if err != nil {
		return sqlcgen.Practicum{}, ErrStudentNotFound
	}
	if student.Role != sqlcgen.UserRoleStudent {
		return sqlcgen.Practicum{}, ErrUserIsNotStudent
	}

	return s.queries.CreatePracticum(ctx, sqlcgen.CreatePracticumParams{
		StudentID:     studentID,
		InstitutionID: institutionID,
		StartDate:     startDate,
		EndDate:       endDate,
	})
}

func (s *Service) CreatePlacement(ctx context.Context, practicumID, agencyID pgtype.UUID, startDate, endDate pgtype.Date) (sqlcgen.Placement, error) {
	if _, err := s.queries.GetPracticumByID(ctx, practicumID); err != nil {
		return sqlcgen.Placement{}, ErrPracticumNotFound
	}

	return s.queries.CreatePlacement(ctx, sqlcgen.CreatePlacementParams{
		PracticumID: practicumID,
		AgencyID:    agencyID,
		StartDate:   startDate,
		EndDate:     endDate,
	})
}

func (s *Service) CreateSupervisorAssignment(ctx context.Context, practicumID, supervisorID pgtype.UUID) (sqlcgen.SupervisorAssignment, error) {
	if _, err := s.queries.GetPracticumByID(ctx, practicumID); err != nil {
		return sqlcgen.SupervisorAssignment{}, ErrPracticumNotFound
	}

	supervisor, err := s.queries.GetUserByID(ctx, supervisorID)
	if err != nil {
		return sqlcgen.SupervisorAssignment{}, ErrSupervisorNotFound
	}
	if supervisor.Role != sqlcgen.UserRoleFacultySupervisor && supervisor.Role != sqlcgen.UserRoleAgencySupervisor {
		return sqlcgen.SupervisorAssignment{}, ErrUserIsNotSupervisor
	}

	exists, err := s.queries.SupervisorAssignmentExists(ctx, sqlcgen.SupervisorAssignmentExistsParams{
		PracticumID:  practicumID,
		SupervisorID: supervisorID,
	})
	if err != nil {
		return sqlcgen.SupervisorAssignment{}, err
	}
	if exists {
		return sqlcgen.SupervisorAssignment{}, ErrAlreadyAssigned
	}

	return s.queries.CreateSupervisorAssignment(ctx, sqlcgen.CreateSupervisorAssignmentParams{
		PracticumID:  practicumID,
		SupervisorID: supervisorID,
	})
}

func (s *Service) GetSummaryForStudent(ctx context.Context, studentID pgtype.UUID) (sqlcgen.GetPracticumSummaryForStudentRow, error) {
	row, err := s.queries.GetPracticumSummaryForStudent(ctx, studentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcgen.GetPracticumSummaryForStudentRow{}, ErrNoActivePracticum
	}
	return row, err
}

func (s *Service) ListForSupervisor(ctx context.Context, supervisorID pgtype.UUID) ([]sqlcgen.ListStudentsForSupervisorRow, error) {
	return s.queries.ListStudentsForSupervisor(ctx, supervisorID)
}

// GetActivePracticumID is used by other Phase 4+ domains (field activities,
// attendance, weekly reports) to scope a student's self-service writes to
// their current practicum, without duplicating this lookup in every
// package.
func (s *Service) GetActivePracticumID(ctx context.Context, studentID pgtype.UUID) (pgtype.UUID, error) {
	id, err := s.queries.GetActivePracticumIDForStudent(ctx, studentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, ErrNoActivePracticum
	}
	return id, err
}
