// Package feedback implements mandatory weekly feedback from both
// supervisors (business requirements §12): one row per (practicum,
// supervisor, week) — both supervisors must provide feedback every weekend.
// This package only records/lists feedback; it does not enforce or alert on
// missed weeks (that's a monitoring/compliance concern, not modeled by the
// business requirements beyond "must provide" — see docs/ARCHITECTURE.md
// for how that's tracked as a future gap, not invented here).
package feedback

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

var ErrNotAssignedSupervisor = errors.New("you are not an assigned supervisor for this student")

type Service struct {
	queries *sqlcgen.Queries
}

func NewService(queries *sqlcgen.Queries) *Service {
	return &Service{queries: queries}
}

func (s *Service) Submit(ctx context.Context, practicumID, supervisorID pgtype.UUID, weekStart pgtype.Date, text string) (sqlcgen.WeeklyFeedback, error) {
	exists, err := s.queries.SupervisorAssignmentExists(ctx, sqlcgen.SupervisorAssignmentExistsParams{
		PracticumID:  practicumID,
		SupervisorID: supervisorID,
	})
	if err != nil {
		return sqlcgen.WeeklyFeedback{}, err
	}
	if !exists {
		return sqlcgen.WeeklyFeedback{}, ErrNotAssignedSupervisor
	}

	return s.queries.CreateWeeklyFeedback(ctx, sqlcgen.CreateWeeklyFeedbackParams{
		PracticumID:   practicumID,
		SupervisorID:  supervisorID,
		WeekStartDate: weekStart,
		Feedback:      text,
	})
}

func (s *Service) ListForStudent(ctx context.Context, studentID pgtype.UUID) ([]sqlcgen.WeeklyFeedback, error) {
	return s.queries.ListFeedbackForStudent(ctx, studentID)
}

func (s *Service) ListFromSupervisor(ctx context.Context, supervisorID pgtype.UUID) ([]sqlcgen.WeeklyFeedback, error) {
	return s.queries.ListFeedbackFromSupervisor(ctx, supervisorID)
}
