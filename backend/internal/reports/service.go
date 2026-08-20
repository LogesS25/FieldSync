// Package reports implements Student weekly report submission (requirements
// FR-07). A report is submitted in a single action — there is no draft
// state, per an explicit scope decision (see docs/ARCHITECTURE.md §5c):
// FR-07 only requires "submit," and a draft/edit workflow isn't specified
// anywhere in the requirements.
package reports

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

func (s *Service) Submit(ctx context.Context, studentID pgtype.UUID, weekStart, weekEnd pgtype.Date, summary string) (sqlcgen.WeeklyReport, error) {
	practicumID, err := s.practicums.GetActivePracticumID(ctx, studentID)
	if err != nil {
		return sqlcgen.WeeklyReport{}, err
	}

	return s.queries.CreateWeeklyReport(ctx, sqlcgen.CreateWeeklyReportParams{
		StudentID:     studentID,
		PracticumID:   practicumID,
		WeekStartDate: weekStart,
		WeekEndDate:   weekEnd,
		Summary:       summary,
	})
}

func (s *Service) ListForStudent(ctx context.Context, studentID pgtype.UUID) ([]sqlcgen.WeeklyReport, error) {
	return s.queries.ListWeeklyReportsForStudent(ctx, studentID)
}
