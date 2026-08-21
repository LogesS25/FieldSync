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

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
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
	queries    *sqlcgen.Queries
	practicums *practicums.Service
	storage    *storage.Storage
}

func NewService(queries *sqlcgen.Queries, practicumsService *practicums.Service, store *storage.Storage) *Service {
	return &Service{queries: queries, practicums: practicumsService, storage: store}
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

	return s.queries.CreateDailyReport(ctx, sqlcgen.CreateDailyReportParams{
		StudentID:        studentID,
		PracticumID:      practicumID,
		ReportDate:       reportDate,
		FilePath:         filePath,
		OriginalFilename: originalFilename,
	})
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

	return s.queries.SetDailyReportAgencyDecision(ctx, sqlcgen.SetDailyReportAgencyDecisionParams{
		ID:               reportID,
		AgencyStatus:     decision,
		AgencyReviewedBy: supervisorID,
	})
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

	return s.queries.SetDailyReportFacultyDecision(ctx, sqlcgen.SetDailyReportFacultyDecisionParams{
		ID:                reportID,
		FacultyStatus:     decision,
		FacultyReviewedBy: supervisorID,
	})
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
