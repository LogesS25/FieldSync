package feedback_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/feedback"
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

func TestSubmit_Success(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	practicumsSvc := practicums.NewService(queries)
	svc := feedback.NewService(queries, notifications.NewService(queries))
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	practicum, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum: %v", err)
	}
	faculty := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)
	if _, err := practicumsSvc.CreateSupervisorAssignment(ctx, practicum.ID, faculty.ID); err != nil {
		t.Fatalf("CreateSupervisorAssignment: %v", err)
	}

	entry, err := svc.Submit(ctx, practicum.ID, faculty.ID, parseDate(t, "2026-01-04"), "Great progress this week.")
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if entry.Feedback != "Great progress this week." {
		t.Errorf("Feedback = %q", entry.Feedback)
	}
}

func TestSubmit_RejectsUnassignedSupervisor(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	practicumsSvc := practicums.NewService(queries)
	svc := feedback.NewService(queries, notifications.NewService(queries))
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	practicum, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum: %v", err)
	}
	unassignedFaculty := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)

	_, err = svc.Submit(ctx, practicum.ID, unassignedFaculty.ID, parseDate(t, "2026-01-04"), "Should be rejected.")
	if !errors.Is(err, feedback.ErrNotAssignedSupervisor) {
		t.Fatalf("error = %v, want ErrNotAssignedSupervisor", err)
	}
}

func TestSubmit_DuplicateForSameWeekRejected(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	practicumsSvc := practicums.NewService(queries)
	svc := feedback.NewService(queries, notifications.NewService(queries))
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	practicum, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum: %v", err)
	}
	faculty := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)
	if _, err := practicumsSvc.CreateSupervisorAssignment(ctx, practicum.ID, faculty.ID); err != nil {
		t.Fatalf("CreateSupervisorAssignment: %v", err)
	}

	if _, err := svc.Submit(ctx, practicum.ID, faculty.ID, parseDate(t, "2026-01-04"), "First."); err != nil {
		t.Fatalf("first Submit returned error: %v", err)
	}
	_, err = svc.Submit(ctx, practicum.ID, faculty.ID, parseDate(t, "2026-01-04"), "Duplicate.")
	if err == nil {
		t.Fatal("expected an error for duplicate (practicum, supervisor, week), got nil")
	}
}

func TestListForStudent_IncludesBothSupervisors(t *testing.T) {
	queries := testutil.NewTestQueries(t)
	practicumsSvc := practicums.NewService(queries)
	svc := feedback.NewService(queries, notifications.NewService(queries))
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	practicum, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum: %v", err)
	}
	faculty := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)
	agencySup := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAgencySupervisor)
	if _, err := practicumsSvc.CreateSupervisorAssignment(ctx, practicum.ID, faculty.ID); err != nil {
		t.Fatalf("assign faculty: %v", err)
	}
	if _, err := practicumsSvc.CreateSupervisorAssignment(ctx, practicum.ID, agencySup.ID); err != nil {
		t.Fatalf("assign agency: %v", err)
	}

	if _, err := svc.Submit(ctx, practicum.ID, faculty.ID, parseDate(t, "2026-01-04"), "Faculty feedback."); err != nil {
		t.Fatalf("Submit(faculty) returned error: %v", err)
	}
	if _, err := svc.Submit(ctx, practicum.ID, agencySup.ID, parseDate(t, "2026-01-04"), "Agency feedback."); err != nil {
		t.Fatalf("Submit(agency) returned error: %v", err)
	}

	entries, err := svc.ListForStudent(ctx, student.ID)
	if err != nil {
		t.Fatalf("ListForStudent returned error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(entries) = %d, want 2", len(entries))
	}
}
