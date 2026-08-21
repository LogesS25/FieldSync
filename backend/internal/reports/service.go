// Package reports implements the "Consolidated Report" (business
// requirements §13): one report per practicum, submitted by the student,
// reviewed agency-then-faculty (each approve/reject), replacing the
// previous unlimited free-text weekly reports with no review — see
// docs/ARCHITECTURE.md §3a #5.
package reports

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/practicums"
)

var (
	ErrNotAssignedSupervisor = errors.New("you are not an assigned supervisor for this student")
	ErrAgencyReviewFirst     = errors.New("agency supervisor must approve before faculty review")
	ErrReportNotFound        = errors.New("consolidated report not found")
	ErrAlreadyReviewedByRole = errors.New("you have already reviewed this report")
)

type Service struct {
	queries    *sqlcgen.Queries
	practicums *practicums.Service
}

func NewService(queries *sqlcgen.Queries, practicumsService *practicums.Service) *Service {
	return &Service{queries: queries, practicums: practicumsService}
}

func (s *Service) Submit(ctx context.Context, studentID pgtype.UUID, summary string) (sqlcgen.ConsolidatedReport, error) {
	practicumID, err := s.practicums.GetActivePracticumID(ctx, studentID)
	if err != nil {
		return sqlcgen.ConsolidatedReport{}, err
	}

	return s.queries.CreateConsolidatedReport(ctx, sqlcgen.CreateConsolidatedReportParams{
		StudentID:   studentID,
		PracticumID: practicumID,
		Summary:     summary,
	})
}

func (s *Service) GetForStudent(ctx context.Context, studentID pgtype.UUID) (sqlcgen.ConsolidatedReport, error) {
	report, err := s.queries.GetConsolidatedReportForStudent(ctx, studentID)
	if err != nil {
		return sqlcgen.ConsolidatedReport{}, ErrReportNotFound
	}
	return report, nil
}

func (s *Service) ListPendingForSupervisor(ctx context.Context, supervisorID pgtype.UUID) ([]sqlcgen.ConsolidatedReport, error) {
	return s.queries.ListPendingConsolidatedReportsForSupervisor(ctx, supervisorID)
}

func (s *Service) AgencyReview(ctx context.Context, reportID, supervisorID pgtype.UUID, approve bool) (sqlcgen.ConsolidatedReport, error) {
	report, err := s.queries.GetConsolidatedReportByID(ctx, reportID)
	if err != nil {
		return sqlcgen.ConsolidatedReport{}, ErrReportNotFound
	}
	if err := s.requireAssignedSupervisor(ctx, report.PracticumID, supervisorID); err != nil {
		return sqlcgen.ConsolidatedReport{}, err
	}
	if report.AgencyStatus != sqlcgen.ReviewDecisionPending {
		return sqlcgen.ConsolidatedReport{}, ErrAlreadyReviewedByRole
	}

	decision := sqlcgen.ReviewDecisionRejected
	if approve {
		decision = sqlcgen.ReviewDecisionApproved
	}

	return s.queries.SetConsolidatedReportAgencyDecision(ctx, sqlcgen.SetConsolidatedReportAgencyDecisionParams{
		ID:               reportID,
		AgencyStatus:     decision,
		AgencyReviewedBy: supervisorID,
	})
}

func (s *Service) FacultyReview(ctx context.Context, reportID, supervisorID pgtype.UUID, approve bool) (sqlcgen.ConsolidatedReport, error) {
	report, err := s.queries.GetConsolidatedReportByID(ctx, reportID)
	if err != nil {
		return sqlcgen.ConsolidatedReport{}, ErrReportNotFound
	}
	if err := s.requireAssignedSupervisor(ctx, report.PracticumID, supervisorID); err != nil {
		return sqlcgen.ConsolidatedReport{}, err
	}
	if report.AgencyStatus != sqlcgen.ReviewDecisionApproved {
		return sqlcgen.ConsolidatedReport{}, ErrAgencyReviewFirst
	}
	if report.FacultyStatus != sqlcgen.ReviewDecisionPending {
		return sqlcgen.ConsolidatedReport{}, ErrAlreadyReviewedByRole
	}

	decision := sqlcgen.ReviewDecisionRejected
	if approve {
		decision = sqlcgen.ReviewDecisionApproved
	}

	return s.queries.SetConsolidatedReportFacultyDecision(ctx, sqlcgen.SetConsolidatedReportFacultyDecisionParams{
		ID:                reportID,
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
