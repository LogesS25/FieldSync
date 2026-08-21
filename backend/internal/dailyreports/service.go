// Package dailyreports implements the student's daily handwritten fieldwork
// report (business requirements §10): a PDF upload per fieldwork day, with
// the same agency-then-faculty sequential review pattern already used by
// attendance_records/consolidated_reports. Correction/resubmission after
// rejection is explicitly TBD in the requirements doc, so — unlike the
// consolidated report — there is deliberately no resubmit here.
package dailyreports

import (
	"context"
	"errors"
	"io"
	"log"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/notifications"
	"github.com/fieldsync/backend/internal/practicums"
	"github.com/fieldsync/backend/internal/storage"
)

var (
	ErrNotAssignedSupervisor = errors.New("you are not an assigned supervisor for this student")
	ErrAgencyReviewFirst     = errors.New("agency supervisor must approve before faculty review")
	ErrReportNotFound        = errors.New("daily report not found")
	ErrAlreadyReviewedByRole = errors.New("you have already reviewed this report")
	ErrNotYourReport         = errors.New("this report does not belong to you")
)

const storageSubdir = "daily-reports"

type Service struct {
	queries       *sqlcgen.Queries
	practicums    *practicums.Service
	storage       *storage.Storage
	notifications *notifications.Service
}

func NewService(queries *sqlcgen.Queries, practicumsService *practicums.Service, store *storage.Storage, notificationsService *notifications.Service) *Service {
	return &Service{queries: queries, practicums: practicumsService, storage: store, notifications: notificationsService}
}

// notify is best-effort: a failed notification insert must never fail the
// business action that triggered it.
func (s *Service) notify(ctx context.Context, recipientID pgtype.UUID, message string) {
	if _, err := s.notifications.Create(ctx, recipientID, message); err != nil {
		log.Printf("dailyreports: failed to create notification for %s: %v", recipientID, err)
	}
}

func (s *Service) Create(ctx context.Context, studentID pgtype.UUID, reportDate pgtype.Date, originalFilename string, file io.Reader) (sqlcgen.DailyReport, error) {
	practicumID, err := s.practicums.GetActivePracticumID(ctx, studentID)
	if err != nil {
		return sqlcgen.DailyReport{}, err
	}

	filePath, err := s.storage.Save(storageSubdir, ".pdf", file)
	if err != nil {
		return sqlcgen.DailyReport{}, err
	}

	report, err := s.queries.CreateDailyReport(ctx, sqlcgen.CreateDailyReportParams{
		StudentID:        studentID,
		PracticumID:      practicumID,
		ReportDate:       reportDate,
		FilePath:         filePath,
		OriginalFilename: originalFilename,
	})
	if err != nil {
		return sqlcgen.DailyReport{}, err
	}

	// Business requirements §10: "the agency supervisor and faculty
	// supervisor are notified" after a daily report is submitted.
	supervisorIDs, err := s.practicums.ListSupervisorIDs(ctx, practicumID)
	if err != nil {
		log.Printf("dailyreports: failed to list supervisors to notify for practicum %s: %v", practicumID, err)
	}
	for _, supervisorID := range supervisorIDs {
		s.notify(ctx, supervisorID, "A student submitted a new daily report awaiting your review.")
	}

	return report, nil
}

func (s *Service) ListForStudent(ctx context.Context, studentID pgtype.UUID) ([]sqlcgen.DailyReport, error) {
	return s.queries.ListDailyReportsForStudent(ctx, studentID)
}

func (s *Service) ListPendingForSupervisor(ctx context.Context, supervisorID pgtype.UUID) ([]sqlcgen.DailyReport, error) {
	return s.queries.ListPendingDailyReportsForSupervisor(ctx, supervisorID)
}

func (s *Service) AgencyReview(ctx context.Context, reportID, supervisorID pgtype.UUID, approve bool) (sqlcgen.DailyReport, error) {
	report, err := s.queries.GetDailyReportByID(ctx, reportID)
	if err != nil {
		return sqlcgen.DailyReport{}, ErrReportNotFound
	}
	if err := s.requireAssignedSupervisor(ctx, report.PracticumID, supervisorID); err != nil {
		return sqlcgen.DailyReport{}, err
	}
	if report.AgencyStatus != sqlcgen.ReviewDecisionPending {
		return sqlcgen.DailyReport{}, ErrAlreadyReviewedByRole
	}

	decision := sqlcgen.ReviewDecisionRejected
	if approve {
		decision = sqlcgen.ReviewDecisionApproved
	}

	updated, err := s.queries.SetDailyReportAgencyDecision(ctx, sqlcgen.SetDailyReportAgencyDecisionParams{
		ID:               reportID,
		AgencyStatus:     decision,
		AgencyReviewedBy: supervisorID,
	})
	if err != nil {
		return sqlcgen.DailyReport{}, err
	}

	s.notify(ctx, updated.StudentID, "Your daily report for "+dateString(updated.ReportDate)+" was "+decisionVerb(approve)+" by your agency supervisor.")
	return updated, nil
}

func (s *Service) FacultyReview(ctx context.Context, reportID, supervisorID pgtype.UUID, approve bool) (sqlcgen.DailyReport, error) {
	report, err := s.queries.GetDailyReportByID(ctx, reportID)
	if err != nil {
		return sqlcgen.DailyReport{}, ErrReportNotFound
	}
	if err := s.requireAssignedSupervisor(ctx, report.PracticumID, supervisorID); err != nil {
		return sqlcgen.DailyReport{}, err
	}
	if report.AgencyStatus != sqlcgen.ReviewDecisionApproved {
		return sqlcgen.DailyReport{}, ErrAgencyReviewFirst
	}
	if report.FacultyStatus != sqlcgen.ReviewDecisionPending {
		return sqlcgen.DailyReport{}, ErrAlreadyReviewedByRole
	}

	decision := sqlcgen.ReviewDecisionRejected
	if approve {
		decision = sqlcgen.ReviewDecisionApproved
	}

	updated, err := s.queries.SetDailyReportFacultyDecision(ctx, sqlcgen.SetDailyReportFacultyDecisionParams{
		ID:                reportID,
		FacultyStatus:     decision,
		FacultyReviewedBy: supervisorID,
	})
	if err != nil {
		return sqlcgen.DailyReport{}, err
	}

	s.notify(ctx, updated.StudentID, "Your daily report for "+dateString(updated.ReportDate)+" was "+decisionVerb(approve)+" by your faculty supervisor.")
	return updated, nil
}

func decisionVerb(approve bool) string {
	if approve {
		return "approved"
	}
	return "rejected"
}

func dateString(d pgtype.Date) string {
	if !d.Valid {
		return "an unknown date"
	}
	return d.Time.Format("2006-01-02")
}

// GetFileForDownload returns the absolute on-disk path and original
// filename for a report, after checking the caller is either the owning
// student or an assigned supervisor on that report's practicum.
func (s *Service) GetFileForDownload(ctx context.Context, reportID, callerID pgtype.UUID, callerIsStudent bool) (absolutePath, filename string, err error) {
	report, getErr := s.queries.GetDailyReportByID(ctx, reportID)
	if getErr != nil {
		return "", "", ErrReportNotFound
	}

	if callerIsStudent {
		if report.StudentID != callerID {
			return "", "", ErrNotYourReport
		}
	} else if assignedErr := s.requireAssignedSupervisor(ctx, report.PracticumID, callerID); assignedErr != nil {
		return "", "", assignedErr
	}

	return s.storage.AbsolutePath(report.FilePath), report.OriginalFilename, nil
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
