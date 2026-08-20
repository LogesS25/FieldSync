package attendance_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/attendance"
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

func newTestSetup(t *testing.T) (*attendance.Service, *practicums.Service, *sqlcgen.Queries) {
	t.Helper()
	queries := testutil.NewTestQueries(t)
	practicumsService := practicums.NewService(queries)
	return attendance.NewService(queries, practicumsService), practicumsService, queries
}

func TestCreate_Success(t *testing.T) {
	svc, practicumsSvc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	if _, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{}); err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	record, err := svc.Create(ctx, student.ID, parseDate(t, "2026-01-05"), 6.5)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got := db.NumericToFloat64(record.HoursLogged); got != 6.5 {
		t.Errorf("HoursLogged = %v, want 6.5", got)
	}
}

func TestCreate_DuplicateDateRejected(t *testing.T) {
	svc, practicumsSvc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	if _, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{}); err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	if _, err := svc.Create(ctx, student.ID, parseDate(t, "2026-01-05"), 4); err != nil {
		t.Fatalf("first Create returned error: %v", err)
	}

	_, err := svc.Create(ctx, student.ID, parseDate(t, "2026-01-05"), 3)
	if err == nil {
		t.Fatal("expected an error for a duplicate attendance_date, got nil")
	}
}

func TestCreate_NoActivePracticum(t *testing.T) {
	svc, _, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	_, err := svc.Create(ctx, student.ID, parseDate(t, "2026-01-05"), 4)
	if !errors.Is(err, practicums.ErrNoActivePracticum) {
		t.Fatalf("error = %v, want ErrNoActivePracticum", err)
	}
}

func TestGetTotalHours_SumsAcrossRecords(t *testing.T) {
	svc, practicumsSvc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	if _, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{}); err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	if _, err := svc.Create(ctx, student.ID, parseDate(t, "2026-01-05"), 6); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := svc.Create(ctx, student.ID, parseDate(t, "2026-01-06"), 4.25); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	total, err := svc.GetTotalHours(ctx, student.ID)
	if err != nil {
		t.Fatalf("GetTotalHours returned error: %v", err)
	}
	if total != 10.25 {
		t.Errorf("total = %v, want 10.25", total)
	}
}

func TestGetTotalHours_NoRecordsReturnsZero(t *testing.T) {
	svc, _, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	total, err := svc.GetTotalHours(ctx, student.ID)
	if err != nil {
		t.Fatalf("GetTotalHours returned error: %v", err)
	}
	if total != 0 {
		t.Errorf("total = %v, want 0", total)
	}
}
