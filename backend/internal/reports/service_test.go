package reports_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/practicums"
	"github.com/fieldsync/backend/internal/reports"
	"github.com/fieldsync/backend/internal/testutil"
)

func parseDate(t *testing.T, s string) pgtype.Date {
	t.Helper()
	d, err := db.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

type testFixture struct {
	svc          *reports.Service
	queries      *sqlcgen.Queries
	student      sqlcgen.User
	faculty      sqlcgen.User
	agencySup    sqlcgen.User
	otherFaculty sqlcgen.User
}

func newFixture(t *testing.T) testFixture {
	t.Helper()
	queries := testutil.NewTestQueries(t)
	practicumsSvc := practicums.NewService(queries)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	practicum, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum: %v", err)
	}

	faculty := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)
	agencySup := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAgencySupervisor)
	otherFaculty := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)

	if _, err := practicumsSvc.CreateSupervisorAssignment(ctx, practicum.ID, faculty.ID); err != nil {
		t.Fatalf("assign faculty: %v", err)
	}
	if _, err := practicumsSvc.CreateSupervisorAssignment(ctx, practicum.ID, agencySup.ID); err != nil {
		t.Fatalf("assign agency supervisor: %v", err)
	}

	return testFixture{
		svc:          reports.NewService(queries, practicumsSvc),
		queries:      queries,
		student:      student,
		faculty:      faculty,
		agencySup:    agencySup,
		otherFaculty: otherFaculty,
	}
}

func TestSubmit_Success(t *testing.T) {
	f := newFixture(t)

	report, err := f.svc.Submit(context.Background(), f.student.ID, "Completed the full fieldwork period.")
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if report.AgencyStatus != sqlcgen.ReviewDecisionPending || report.FacultyStatus != sqlcgen.ReviewDecisionPending {
		t.Errorf("expected both statuses pending, got agency=%v faculty=%v", report.AgencyStatus, report.FacultyStatus)
	}
}

func TestSubmit_SecondReportForSamePracticumRejected(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	if _, err := f.svc.Submit(ctx, f.student.ID, "First."); err != nil {
		t.Fatalf("first Submit returned error: %v", err)
	}
	_, err := f.svc.Submit(ctx, f.student.ID, "Second.")
	if err == nil {
		t.Fatal("expected an error submitting a second consolidated report for the same practicum, got nil")
	}
}

func TestSubmit_NoActivePracticum(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	svc := reports.NewService(queries, practicums.NewService(queries))
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	_, err := svc.Submit(context.Background(), student.ID, "No practicum yet.")
	if !errors.Is(err, practicums.ErrNoActivePracticum) {
		t.Fatalf("error = %v, want ErrNoActivePracticum", err)
	}
}

func TestFacultyReview_RejectedBeforeAgencyApproval(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Submit(ctx, f.student.ID, "Ready for review.")
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}

	_, err = f.svc.FacultyReview(ctx, report.ID, f.faculty.ID, true)
	if !errors.Is(err, reports.ErrAgencyReviewFirst) {
		t.Fatalf("error = %v, want ErrAgencyReviewFirst", err)
	}
}

func TestSequentialApproval_AgencyThenFaculty(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Submit(ctx, f.student.ID, "Ready for review.")
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}

	if _, err := f.svc.AgencyReview(ctx, report.ID, f.agencySup.ID, true); err != nil {
		t.Fatalf("AgencyReview returned error: %v", err)
	}

	afterFaculty, err := f.svc.FacultyReview(ctx, report.ID, f.faculty.ID, true)
	if err != nil {
		t.Fatalf("FacultyReview returned error: %v", err)
	}
	if afterFaculty.FacultyStatus != sqlcgen.ReviewDecisionApproved {
		t.Errorf("FacultyStatus = %v, want approved", afterFaculty.FacultyStatus)
	}
}

func TestAgencyReview_RejectsUnassignedSupervisor(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Submit(ctx, f.student.ID, "Ready for review.")
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}

	_, err = f.svc.AgencyReview(ctx, report.ID, f.otherFaculty.ID, true)
	if !errors.Is(err, reports.ErrNotAssignedSupervisor) {
		t.Fatalf("error = %v, want ErrNotAssignedSupervisor", err)
	}
}

func TestGetForStudent_NotFound(t *testing.T) {
	f := newFixture(t)

	_, err := f.svc.GetForStudent(context.Background(), f.student.ID)
	if !errors.Is(err, reports.ErrReportNotFound) {
		t.Fatalf("error = %v, want ErrReportNotFound", err)
	}
}

func TestResubmit_RejectsWhenNotYetReviewed(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Submit(ctx, f.student.ID, "First submission.")
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}

	_, err = f.svc.Resubmit(ctx, report.ID, f.student.ID, "Trying to resubmit early.")
	if !errors.Is(err, reports.ErrNotRejected) {
		t.Fatalf("error = %v, want ErrNotRejected", err)
	}
}

func TestResubmit_AfterRejectionGoesThroughApprovalAgain(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Submit(ctx, f.student.ID, "First submission.")
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if _, err := f.svc.AgencyReview(ctx, report.ID, f.agencySup.ID, false); err != nil {
		t.Fatalf("AgencyReview returned error: %v", err)
	}

	resubmitted, err := f.svc.Resubmit(ctx, report.ID, f.student.ID, "Revised submission.")
	if err != nil {
		t.Fatalf("Resubmit returned error: %v", err)
	}
	if resubmitted.AgencyStatus != sqlcgen.ReviewDecisionPending {
		t.Errorf("AgencyStatus = %v, want pending (must go through approval again)", resubmitted.AgencyStatus)
	}
	if resubmitted.Summary != "Revised submission." {
		t.Errorf("Summary = %q, want the resubmitted text", resubmitted.Summary)
	}

	// The resubmitted report should now behave exactly like a fresh
	// submission for review purposes.
	if _, err := f.svc.AgencyReview(ctx, report.ID, f.agencySup.ID, true); err != nil {
		t.Fatalf("AgencyReview after resubmit returned error: %v", err)
	}
}

func TestResubmit_RejectsWrongStudent(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Submit(ctx, f.student.ID, "First submission.")
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if _, err := f.svc.AgencyReview(ctx, report.ID, f.agencySup.ID, false); err != nil {
		t.Fatalf("AgencyReview returned error: %v", err)
	}

	otherStudent := testutil.CreateTestUser(t, f.queries, sqlcgen.UserRoleStudent)
	_, err = f.svc.Resubmit(ctx, report.ID, otherStudent.ID, "Should be rejected.")
	if !errors.Is(err, reports.ErrNotYourReport) {
		t.Fatalf("error = %v, want ErrNotYourReport", err)
	}
}
