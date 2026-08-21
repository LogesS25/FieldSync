package attendance_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/attendance"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/notifications"
	"github.com/fieldsync/backend/internal/practicums"
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

// testFixture is a student with an active practicum and one assigned
// faculty + one assigned agency supervisor — the setup every sequential-
// approval test needs.
type testFixture struct {
	svc          *attendance.Service
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
		svc:          attendance.NewService(queries, practicumsSvc, notifications.NewService(queries)),
		queries:      queries,
		student:      student,
		faculty:      faculty,
		agencySup:    agencySup,
		otherFaculty: otherFaculty,
	}
}

func TestCreate_Success(t *testing.T) {
	f := newFixture(t)
	hours := 6.5

	record, err := f.svc.Create(context.Background(), attendance.CreateInput{
		StudentID:      f.student.ID,
		AttendanceDate: parseDate(t, "2026-01-05"),
		Session:        sqlcgen.AttendanceSessionMorning,
		Hours:          &hours,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if record.AgencyStatus != sqlcgen.ReviewDecisionPending {
		t.Errorf("AgencyStatus = %v, want pending", record.AgencyStatus)
	}
	if record.FacultyStatus != sqlcgen.ReviewDecisionPending {
		t.Errorf("FacultyStatus = %v, want pending", record.FacultyStatus)
	}
}

func TestCreate_SameDateDifferentSessionAllowed(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	if _, err := f.svc.Create(ctx, attendance.CreateInput{
		StudentID: f.student.ID, AttendanceDate: parseDate(t, "2026-01-05"), Session: sqlcgen.AttendanceSessionMorning,
	}); err != nil {
		t.Fatalf("Create(morning) returned error: %v", err)
	}
	if _, err := f.svc.Create(ctx, attendance.CreateInput{
		StudentID: f.student.ID, AttendanceDate: parseDate(t, "2026-01-05"), Session: sqlcgen.AttendanceSessionEvening,
	}); err != nil {
		t.Fatalf("Create(evening) returned error: %v", err)
	}
}

func TestCreate_SameDateSameSessionRejected(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	if _, err := f.svc.Create(ctx, attendance.CreateInput{
		StudentID: f.student.ID, AttendanceDate: parseDate(t, "2026-01-05"), Session: sqlcgen.AttendanceSessionMorning,
	}); err != nil {
		t.Fatalf("first Create returned error: %v", err)
	}
	_, err := f.svc.Create(ctx, attendance.CreateInput{
		StudentID: f.student.ID, AttendanceDate: parseDate(t, "2026-01-05"), Session: sqlcgen.AttendanceSessionMorning,
	})
	if err == nil {
		t.Fatal("expected an error for duplicate (date, session), got nil")
	}
}

func TestFacultyReview_RejectedBeforeAgencyApproval(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	record, err := f.svc.Create(ctx, attendance.CreateInput{
		StudentID: f.student.ID, AttendanceDate: parseDate(t, "2026-01-05"), Session: sqlcgen.AttendanceSessionMorning,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, err = f.svc.FacultyReview(ctx, record.ID, f.faculty.ID, true)
	if !errors.Is(err, attendance.ErrAgencyReviewFirst) {
		t.Fatalf("error = %v, want ErrAgencyReviewFirst", err)
	}
}

func TestSequentialApproval_AgencyThenFaculty(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	record, err := f.svc.Create(ctx, attendance.CreateInput{
		StudentID: f.student.ID, AttendanceDate: parseDate(t, "2026-01-05"), Session: sqlcgen.AttendanceSessionMorning,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	afterAgency, err := f.svc.AgencyReview(ctx, record.ID, f.agencySup.ID, true)
	if err != nil {
		t.Fatalf("AgencyReview returned error: %v", err)
	}
	if afterAgency.AgencyStatus != sqlcgen.ReviewDecisionApproved {
		t.Fatalf("AgencyStatus = %v, want approved", afterAgency.AgencyStatus)
	}

	afterFaculty, err := f.svc.FacultyReview(ctx, record.ID, f.faculty.ID, true)
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

	record, err := f.svc.Create(ctx, attendance.CreateInput{
		StudentID: f.student.ID, AttendanceDate: parseDate(t, "2026-01-05"), Session: sqlcgen.AttendanceSessionMorning,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, err = f.svc.AgencyReview(ctx, record.ID, f.otherFaculty.ID, true)
	if !errors.Is(err, attendance.ErrNotAssignedSupervisor) {
		t.Fatalf("error = %v, want ErrNotAssignedSupervisor", err)
	}
}

func TestAgencyReview_CannotReviewTwice(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	record, err := f.svc.Create(ctx, attendance.CreateInput{
		StudentID: f.student.ID, AttendanceDate: parseDate(t, "2026-01-05"), Session: sqlcgen.AttendanceSessionMorning,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if _, err := f.svc.AgencyReview(ctx, record.ID, f.agencySup.ID, true); err != nil {
		t.Fatalf("first AgencyReview returned error: %v", err)
	}
	_, err = f.svc.AgencyReview(ctx, record.ID, f.agencySup.ID, false)
	if !errors.Is(err, attendance.ErrAlreadyReviewedByRole) {
		t.Fatalf("error = %v, want ErrAlreadyReviewedByRole", err)
	}
}

func TestListPendingForSupervisor_AgencyBeforeApprovalFacultyAfter(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	record, err := f.svc.Create(ctx, attendance.CreateInput{
		StudentID: f.student.ID, AttendanceDate: parseDate(t, "2026-01-05"), Session: sqlcgen.AttendanceSessionMorning,
	})
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

	if _, err := f.svc.AgencyReview(ctx, record.ID, f.agencySup.ID, true); err != nil {
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

func TestCreate_NoActivePracticum(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	svc := attendance.NewService(queries, practicums.NewService(queries), notifications.NewService(queries))
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	_, err := svc.Create(context.Background(), attendance.CreateInput{
		StudentID: student.ID, AttendanceDate: parseDate(t, "2026-01-05"), Session: sqlcgen.AttendanceSessionMorning,
	})
	if !errors.Is(err, practicums.ErrNoActivePracticum) {
		t.Fatalf("error = %v, want ErrNoActivePracticum", err)
	}
}
