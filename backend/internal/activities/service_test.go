package activities_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/activities"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
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

func newTestSetup(t *testing.T) (*activities.Service, *practicums.Service, *sqlcgen.Queries) {
	t.Helper()
	queries := testutil.NewTestQueries(t)
	practicumsService := practicums.NewService(queries)
	return activities.NewService(queries, practicumsService), practicumsService, queries
}

func TestCreate_Success(t *testing.T) {
	svc, practicumsSvc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	if _, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{}); err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	activity, err := svc.Create(ctx, student.ID, parseDate(t, "2026-01-05"), "Conducted an intake interview.")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if activity.Description != "Conducted an intake interview." {
		t.Errorf("Description = %q", activity.Description)
	}
	if activity.VerificationStatus != sqlcgen.VerificationStatusPending {
		t.Errorf("VerificationStatus = %v, want pending", activity.VerificationStatus)
	}
}

func TestCreate_NoActivePracticum(t *testing.T) {
	svc, _, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	_, err := svc.Create(ctx, student.ID, parseDate(t, "2026-01-05"), "No practicum yet.")
	if !errors.Is(err, practicums.ErrNoActivePracticum) {
		t.Fatalf("error = %v, want ErrNoActivePracticum", err)
	}
}

func TestListForStudent_ScopedToOwnRecords(t *testing.T) {
	svc, practicumsSvc, queries := newTestSetup(t)
	ctx := context.Background()

	studentA := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	studentB := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)

	if _, err := practicumsSvc.CreatePracticum(ctx, studentA.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{}); err != nil {
		t.Fatalf("CreatePracticum(A) returned error: %v", err)
	}
	if _, err := practicumsSvc.CreatePracticum(ctx, studentB.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{}); err != nil {
		t.Fatalf("CreatePracticum(B) returned error: %v", err)
	}

	if _, err := svc.Create(ctx, studentA.ID, parseDate(t, "2026-01-05"), "Student A's activity"); err != nil {
		t.Fatalf("Create(A) returned error: %v", err)
	}
	if _, err := svc.Create(ctx, studentB.ID, parseDate(t, "2026-01-05"), "Student B's activity"); err != nil {
		t.Fatalf("Create(B) returned error: %v", err)
	}

	listA, err := svc.ListForStudent(ctx, studentA.ID)
	if err != nil {
		t.Fatalf("ListForStudent(A) returned error: %v", err)
	}
	if len(listA) != 1 || listA[0].Description != "Student A's activity" {
		t.Fatalf("listA = %+v, want exactly student A's own activity", listA)
	}
}
