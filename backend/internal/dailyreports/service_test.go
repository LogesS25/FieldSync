package dailyreports_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/dailyreports"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/notifications"
	"github.com/fieldsync/backend/internal/practicums"
	"github.com/fieldsync/backend/internal/storage"
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
	svc          *dailyreports.Service
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

	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}

	return testFixture{
		svc:          dailyreports.NewService(queries, practicumsSvc, store, notifications.NewService(queries)),
		queries:      queries,
		student:      student,
		faculty:      faculty,
		agencySup:    agencySup,
		otherFaculty: otherFaculty,
	}
}

func TestCreate_Success(t *testing.T) {
	f := newFixture(t)

	report, err := f.svc.Create(context.Background(), f.student.ID, parseDate(t, "2026-01-05"), "report.pdf", strings.NewReader("%PDF-1.4 fake"))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if report.AgencyStatus != sqlcgen.ReviewDecisionPending {
		t.Errorf("AgencyStatus = %v, want pending", report.AgencyStatus)
	}
	if report.FacultyStatus != sqlcgen.ReviewDecisionPending {
		t.Errorf("FacultyStatus = %v, want pending", report.FacultyStatus)
	}
	if report.OriginalFilename != "report.pdf" {
		t.Errorf("OriginalFilename = %q, want %q", report.OriginalFilename, "report.pdf")
	}
}

func TestCreate_DuplicateDateRejected(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	if _, err := f.svc.Create(ctx, f.student.ID, parseDate(t, "2026-01-05"), "a.pdf", strings.NewReader("a")); err != nil {
		t.Fatalf("first Create returned error: %v", err)
	}
	_, err := f.svc.Create(ctx, f.student.ID, parseDate(t, "2026-01-05"), "b.pdf", strings.NewReader("b"))
	if err == nil {
		t.Fatal("expected an error for duplicate report date, got nil")
	}
}

func TestCreate_NoActivePracticum(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	svc := dailyreports.NewService(queries, practicums.NewService(queries), store, notifications.NewService(queries))
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	_, err = svc.Create(context.Background(), student.ID, parseDate(t, "2026-01-05"), "a.pdf", strings.NewReader("a"))
	if !errors.Is(err, practicums.ErrNoActivePracticum) {
		t.Fatalf("error = %v, want ErrNoActivePracticum", err)
	}
}

func TestFacultyReview_RejectedBeforeAgencyApproval(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Create(ctx, f.student.ID, parseDate(t, "2026-01-05"), "a.pdf", strings.NewReader("a"))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, err = f.svc.FacultyReview(ctx, report.ID, f.faculty.ID, true)
	if !errors.Is(err, dailyreports.ErrAgencyReviewFirst) {
		t.Fatalf("error = %v, want ErrAgencyReviewFirst", err)
	}
}

func TestSequentialApproval_AgencyThenFaculty(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Create(ctx, f.student.ID, parseDate(t, "2026-01-05"), "a.pdf", strings.NewReader("a"))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	afterAgency, err := f.svc.AgencyReview(ctx, report.ID, f.agencySup.ID, true)
	if err != nil {
		t.Fatalf("AgencyReview returned error: %v", err)
	}
	if afterAgency.AgencyStatus != sqlcgen.ReviewDecisionApproved {
		t.Fatalf("AgencyStatus = %v, want approved", afterAgency.AgencyStatus)
	}

	afterFaculty, err := f.svc.FacultyReview(ctx, report.ID, f.faculty.ID, true)
	if err != nil {
		t.Fatalf("FacultyReview returned error: %v", err)
	}
	if afterFaculty.FacultyStatus != sqlcgen.ReviewDecisionApproved {
		t.Fatalf("FacultyStatus = %v, want approved", afterFaculty.FacultyStatus)
	}
}

func TestAgencyReview_RejectsUnassignedSupervisor(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Create(ctx, f.student.ID, parseDate(t, "2026-01-05"), "a.pdf", strings.NewReader("a"))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, err = f.svc.AgencyReview(ctx, report.ID, f.otherFaculty.ID, true)
	if !errors.Is(err, dailyreports.ErrNotAssignedSupervisor) {
		t.Fatalf("error = %v, want ErrNotAssignedSupervisor", err)
	}
}

func TestAgencyReview_CannotReviewTwice(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Create(ctx, f.student.ID, parseDate(t, "2026-01-05"), "a.pdf", strings.NewReader("a"))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if _, err := f.svc.AgencyReview(ctx, report.ID, f.agencySup.ID, true); err != nil {
		t.Fatalf("first AgencyReview returned error: %v", err)
	}
	_, err = f.svc.AgencyReview(ctx, report.ID, f.agencySup.ID, false)
	if !errors.Is(err, dailyreports.ErrAlreadyReviewedByRole) {
		t.Fatalf("error = %v, want ErrAlreadyReviewedByRole", err)
	}
}

func TestListPendingForSupervisor_AgencyBeforeApprovalFacultyAfter(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Create(ctx, f.student.ID, parseDate(t, "2026-01-05"), "a.pdf", strings.NewReader("a"))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	agencyPending, err := f.svc.ListPendingForSupervisor(ctx, f.agencySup.ID)
	if err != nil {
		t.Fatalf("ListPendingForSupervisor(agency) returned error: %v", err)
	}
	if len(agencyPending) != 1 {
		t.Fatalf("len(agencyPending) = %d, want 1", len(agencyPending))
	}

	facultyPendingBefore, err := f.svc.ListPendingForSupervisor(ctx, f.faculty.ID)
	if err != nil {
		t.Fatalf("ListPendingForSupervisor(faculty, before) returned error: %v", err)
	}
	if len(facultyPendingBefore) != 0 {
		t.Fatalf("len(facultyPendingBefore) = %d, want 0 (agency hasn't approved yet)", len(facultyPendingBefore))
	}

	if _, err := f.svc.AgencyReview(ctx, report.ID, f.agencySup.ID, true); err != nil {
		t.Fatalf("AgencyReview returned error: %v", err)
	}

	facultyPendingAfter, err := f.svc.ListPendingForSupervisor(ctx, f.faculty.ID)
	if err != nil {
		t.Fatalf("ListPendingForSupervisor(faculty, after) returned error: %v", err)
	}
	if len(facultyPendingAfter) != 1 {
		t.Fatalf("len(facultyPendingAfter) = %d, want 1 (now that agency approved)", len(facultyPendingAfter))
	}
}

func TestGetFileForDownload_StudentCanReadOwnReport(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Create(ctx, f.student.ID, parseDate(t, "2026-01-05"), "a.pdf", strings.NewReader("a"))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, filename, err := f.svc.GetFileForDownload(ctx, report.ID, f.student.ID, true)
	if err != nil {
		t.Fatalf("GetFileForDownload returned error: %v", err)
	}
	if filename != "a.pdf" {
		t.Errorf("filename = %q, want %q", filename, "a.pdf")
	}
}

func TestGetFileForDownload_OtherStudentForbidden(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Create(ctx, f.student.ID, parseDate(t, "2026-01-05"), "a.pdf", strings.NewReader("a"))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	otherStudent := testutil.CreateTestUser(t, f.queries, sqlcgen.UserRoleStudent)
	_, _, err = f.svc.GetFileForDownload(ctx, report.ID, otherStudent.ID, true)
	if !errors.Is(err, dailyreports.ErrNotYourReport) {
		t.Fatalf("error = %v, want ErrNotYourReport", err)
	}
}

func TestGetFileForDownload_UnassignedSupervisorForbidden(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	report, err := f.svc.Create(ctx, f.student.ID, parseDate(t, "2026-01-05"), "a.pdf", strings.NewReader("a"))
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, _, err = f.svc.GetFileForDownload(ctx, report.ID, f.otherFaculty.ID, false)
	if !errors.Is(err, dailyreports.ErrNotAssignedSupervisor) {
		t.Fatalf("error = %v, want ErrNotAssignedSupervisor", err)
	}
}
