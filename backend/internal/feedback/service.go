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
	"log"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/notifications"
)

var ErrNotAssignedSupervisor = errors.New("you are not an assigned supervisor for this student")

type Service struct {
	queries       *sqlcgen.Queries
	notifications *notifications.Service
}

func NewService(queries *sqlcgen.Queries, notificationsService *notifications.Service) *Service {
	return &Service{queries: queries, notifications: notificationsService}
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

	created, err := s.queries.CreateWeeklyFeedback(ctx, sqlcgen.CreateWeeklyFeedbackParams{
		PracticumID:   practicumID,
		SupervisorID:  supervisorID,
		WeekStartDate: weekStart,
		Feedback:      text,
	})
	if err != nil {
		return sqlcgen.WeeklyFeedback{}, err
	}

	practicum, err := s.queries.GetPracticumByID(ctx, practicumID)
	if err != nil {
		log.Printf("feedback: failed to look up practicum %s to notify student: %v", practicumID, err)
		return created, nil
	}
	if _, err := s.notifications.Create(ctx, practicum.StudentID, "You received new weekly feedback from a supervisor."); err != nil {
		log.Printf("feedback: failed to create notification for %s: %v", practicum.StudentID, err)
	}

	return created, nil
}

func (s *Service) ListForStudent(ctx context.Context, studentID pgtype.UUID) ([]sqlcgen.WeeklyFeedback, error) {
	return s.queries.ListFeedbackForStudent(ctx, studentID)
}

func (s *Service) ListFromSupervisor(ctx context.Context, supervisorID pgtype.UUID) ([]sqlcgen.WeeklyFeedback, error) {
	return s.queries.ListFeedbackFromSupervisor(ctx, supervisorID)
}
