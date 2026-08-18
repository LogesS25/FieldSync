package practicums_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/practicums"
	"github.com/fieldsync/backend/internal/testutil"
)

func newTestSetup(t *testing.T) (*practicums.Service, *sqlcgen.Queries) {
	t.Helper()
	queries := testutil.NewTestQueries(t)
	return practicums.NewService(queries), queries
}

func parseDate(t *testing.T, s string) pgtype.Date {
	t.Helper()
	d, err := db.ParseDate(s)
	if err != nil {
		t.Fatalf("ParseDate(%q): %v", s, err)
	}
	return d
}

func TestCreatePracticum_Success(t *testing.T) {
	svc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)

	practicum, err := svc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}
	if practicum.StudentID != student.ID {
		t.Error("StudentID mismatch")
	}
	if practicum.Status != sqlcgen.PracticumStatusActive {
		t.Errorf("Status = %v, want active", practicum.Status)
	}
}

func TestCreatePracticum_StudentNotFound(t *testing.T) {
	svc, queries := newTestSetup(t)
	ctx := context.Background()

	institution := testutil.CreateTestInstitution(t, queries)
	fakeStudentID, err := db.ParseUUID("00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("ParseUUID: %v", err)
	}

	_, err = svc.CreatePracticum(ctx, fakeStudentID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if !errors.Is(err, practicums.ErrStudentNotFound) {
		t.Fatalf("error = %v, want ErrStudentNotFound", err)
	}
}

func TestCreatePracticum_RejectsNonStudentRole(t *testing.T) {
	svc, queries := newTestSetup(t)
	ctx := context.Background()

	notAStudent := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)
	institution := testutil.CreateTestInstitution(t, queries)

	_, err := svc.CreatePracticum(ctx, notAStudent.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if !errors.Is(err, practicums.ErrUserIsNotStudent) {
		t.Fatalf("error = %v, want ErrUserIsNotStudent", err)
	}
}

func TestCreatePlacement_Success(t *testing.T) {
	svc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	agency := testutil.CreateTestAgency(t, queries)

	practicum, err := svc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	placement, err := svc.CreatePlacement(ctx, practicum.ID, agency.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePlacement returned error: %v", err)
	}
	if placement.AgencyID != agency.ID {
		t.Error("AgencyID mismatch")
	}
}

func TestCreatePlacement_PracticumNotFound(t *testing.T) {
	svc, queries := newTestSetup(t)
	ctx := context.Background()

	agency := testutil.CreateTestAgency(t, queries)
	fakePracticumID, err := db.ParseUUID("00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("ParseUUID: %v", err)
	}

	_, err = svc.CreatePlacement(ctx, fakePracticumID, agency.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if !errors.Is(err, practicums.ErrPracticumNotFound) {
		t.Fatalf("error = %v, want ErrPracticumNotFound", err)
	}
}

func TestCreateSupervisorAssignment_Success(t *testing.T) {
	svc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	supervisor := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)

	practicum, err := svc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	assignment, err := svc.CreateSupervisorAssignment(ctx, practicum.ID, supervisor.ID)
	if err != nil {
		t.Fatalf("CreateSupervisorAssignment returned error: %v", err)
	}
	if assignment.SupervisorID != supervisor.ID {
		t.Error("SupervisorID mismatch")
	}
}

func TestCreateSupervisorAssignment_RejectsStudentAsSupervisor(t *testing.T) {
	svc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	anotherStudent := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	practicum, err := svc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	_, err = svc.CreateSupervisorAssignment(ctx, practicum.ID, anotherStudent.ID)
	if !errors.Is(err, practicums.ErrUserIsNotSupervisor) {
		t.Fatalf("error = %v, want ErrUserIsNotSupervisor", err)
	}
}

func TestCreateSupervisorAssignment_DuplicateRejected(t *testing.T) {
	svc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	supervisor := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAgencySupervisor)

	practicum, err := svc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	if _, err := svc.CreateSupervisorAssignment(ctx, practicum.ID, supervisor.ID); err != nil {
		t.Fatalf("first CreateSupervisorAssignment returned error: %v", err)
	}

	_, err = svc.CreateSupervisorAssignment(ctx, practicum.ID, supervisor.ID)
	if !errors.Is(err, practicums.ErrAlreadyAssigned) {
		t.Fatalf("error = %v, want ErrAlreadyAssigned", err)
	}
}

func TestGetSummaryForStudent_NoActivePracticum(t *testing.T) {
	svc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	_, err := svc.GetSummaryForStudent(ctx, student.ID)
	if !errors.Is(err, practicums.ErrNoActivePracticum) {
		t.Fatalf("error = %v, want ErrNoActivePracticum", err)
	}
}

func TestGetSummaryForStudent_WithPlacementAndSupervisors(t *testing.T) {
	svc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	agency := testutil.CreateTestAgency(t, queries)
	supervisor := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)

	practicum, err := svc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}
	if _, err := svc.CreatePlacement(ctx, practicum.ID, agency.ID, parseDate(t, "2026-01-01"), pgtype.Date{}); err != nil {
		t.Fatalf("CreatePlacement returned error: %v", err)
	}
	if _, err := svc.CreateSupervisorAssignment(ctx, practicum.ID, supervisor.ID); err != nil {
		t.Fatalf("CreateSupervisorAssignment returned error: %v", err)
	}

	summary, err := svc.GetSummaryForStudent(ctx, student.ID)
	if err != nil {
		t.Fatalf("GetSummaryForStudent returned error: %v", err)
	}
	if summary.InstitutionName != institution.Name {
		t.Errorf("InstitutionName = %q, want %q", summary.InstitutionName, institution.Name)
	}
	if !summary.AgencyName.Valid || summary.AgencyName.String != agency.Name {
		t.Errorf("AgencyName = %+v, want %q", summary.AgencyName, agency.Name)
	}
	if len(summary.Supervisors) == 0 {
		t.Error("expected non-empty supervisors JSON array")
	}
}

func TestListForSupervisor_ReturnsAssignedStudents(t *testing.T) {
	svc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	supervisor := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)
	unrelatedSupervisor := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)

	practicum, err := svc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}
	if _, err := svc.CreateSupervisorAssignment(ctx, practicum.ID, supervisor.ID); err != nil {
		t.Fatalf("CreateSupervisorAssignment returned error: %v", err)
	}

	students, err := svc.ListForSupervisor(ctx, supervisor.ID)
	if err != nil {
		t.Fatalf("ListForSupervisor returned error: %v", err)
	}
	if len(students) != 1 {
		t.Fatalf("len(students) = %d, want 1", len(students))
	}
	if students[0].StudentID != student.ID {
		t.Error("StudentID mismatch")
	}

	unrelatedStudents, err := svc.ListForSupervisor(ctx, unrelatedSupervisor.ID)
	if err != nil {
		t.Fatalf("ListForSupervisor(unrelated) returned error: %v", err)
	}
	if len(unrelatedStudents) != 0 {
		t.Fatalf("len(unrelatedStudents) = %d, want 0 (must not see other supervisors' students)", len(unrelatedStudents))
	}
}
