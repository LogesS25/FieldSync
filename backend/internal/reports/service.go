// Package reports implements the "Consolidated Report" (business
// requirements §13): one report per practicum, submitted by the student,
// reviewed agency-then-faculty (each approve/reject), replacing the
// previous unlimited free-text weekly reports with no review — see
// docs/ARCHITECTURE.md §3a #5.
package reports

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/notifications"
	"github.com/fieldsync/backend/internal/practicums"
)

var (
	ErrNotAssignedSupervisor = errors.New("you are not an assigned supervisor for this student")
	ErrAgencyReviewFirst     = errors.New("agency supervisor must approve before faculty review")
	ErrReportNotFound        = errors.New("consolidated report not found")
	ErrAlreadyReviewedByRole = errors.New("you have already reviewed this report")
	ErrNotYourReport         = errors.New("this is not your report")
	ErrNotRejected           = errors.New("only a rejected report can be resubmitted")
)

type Service struct {
	queries       *sqlcgen.Queries
	practicums    *practicums.Service
	notifications *notifications.Service
}

func NewService(queries *sqlcgen.Queries, practicumsService *practicums.Service, notificationsService *notifications.Service) *Service {
	return &Service{queries: queries, practicums: practicumsService, notifications: notificationsService}
}

// notify is best-effort: a failed notification insert must never fail the
// business action that triggered it.
func (s *Service) notify(ctx context.Context, recipientID pgtype.UUID, message string) {
	if _, err := s.notifications.Create(ctx, recipientID, message); err != nil {
		log.Printf("reports: failed to create notification for %s: %v", recipientID, err)
	}
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

// Resubmit implements the business requirements' "if rejected, the student
// must resubmit; the resubmitted report goes through the approval process
// again" (§13) — both review decisions reset to pending, not a fast-tracked
// re-approval of the fixed content.
func (s *Service) Resubmit(ctx context.Context, reportID, studentID pgtype.UUID, summary string) (sqlcgen.ConsolidatedReport, error) {
	report, err := s.queries.GetConsolidatedReportByID(ctx, reportID)
	if err != nil {
		return sqlcgen.ConsolidatedReport{}, ErrReportNotFound
	}
	if report.StudentID != studentID {
		return sqlcgen.ConsolidatedReport{}, ErrNotYourReport
	}
	if report.AgencyStatus != sqlcgen.ReviewDecisionRejected && report.FacultyStatus != sqlcgen.ReviewDecisionRejected {
		return sqlcgen.ConsolidatedReport{}, ErrNotRejected
	}

	updated, err := s.queries.ResubmitConsolidatedReport(ctx, sqlcgen.ResubmitConsolidatedReportParams{
		ID:      reportID,
		Summary: summary,
	})
	if err != nil {
		return sqlcgen.ConsolidatedReport{}, err
	}

	supervisorIDs, listErr := s.practicums.ListSupervisorIDs(ctx, updated.PracticumID)
	if listErr != nil {
		log.Printf("reports: failed to list supervisors to notify for practicum %s: %v", updated.PracticumID, listErr)
	}
	for _, supervisorID := range supervisorIDs {
		s.notify(ctx, supervisorID, "A student resubmitted their consolidated report for review.")
	}

	return updated, nil
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

	updated, err := s.queries.SetConsolidatedReportAgencyDecision(ctx, sqlcgen.SetConsolidatedReportAgencyDecisionParams{
		ID:               reportID,
		AgencyStatus:     decision,
		AgencyReviewedBy: supervisorID,
	})
	if err != nil {
		return sqlcgen.ConsolidatedReport{}, err
	}

	s.notify(ctx, updated.StudentID, "Your consolidated report was "+decisionVerb(approve)+" by your agency supervisor.")
	return updated, nil
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

	updated, err := s.queries.SetConsolidatedReportFacultyDecision(ctx, sqlcgen.SetConsolidatedReportFacultyDecisionParams{
		ID:                reportID,
		FacultyStatus:     decision,
		FacultyReviewedBy: supervisorID,
	})
	if err != nil {
		return sqlcgen.ConsolidatedReport{}, err
	}

	s.notify(ctx, updated.StudentID, "Your consolidated report was "+decisionVerb(approve)+" by your faculty supervisor.")
	return updated, nil
}

func decisionVerb(approve bool) string {
	if approve {
		return "approved"
	}
	return "rejected"
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
