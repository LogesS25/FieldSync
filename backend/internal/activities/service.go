// Package activities implements the Student's daily field activity log
// (requirements FR-05). Records are write-once from the student's side —
// editing/deleting after creation is an open product decision (see
// docs/ARCHITECTURE.md §8) and is deliberately not implemented yet.
package activities

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/practicums"
)

type Service struct {
	queries    *sqlcgen.Queries
	practicums *practicums.Service
}

func NewService(queries *sqlcgen.Queries, practicumsService *practicums.Service) *Service {
	return &Service{queries: queries, practicums: practicumsService}
}

func (s *Service) Create(ctx context.Context, studentID pgtype.UUID, activityDate pgtype.Date, description string) (sqlcgen.FieldActivity, error) {
	practicumID, err := s.practicums.GetActivePracticumID(ctx, studentID)
	if err != nil {
		return sqlcgen.FieldActivity{}, err
	}

	return s.queries.CreateFieldActivity(ctx, sqlcgen.CreateFieldActivityParams{
		StudentID:    studentID,
		PracticumID:  practicumID,
		ActivityDate: activityDate,
		Description:  description,
	})
}

func (s *Service) ListForStudent(ctx context.Context, studentID pgtype.UUID) ([]sqlcgen.FieldActivity, error) {
	return s.queries.ListFieldActivitiesForStudent(ctx, studentID)
}
