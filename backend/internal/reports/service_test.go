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

func newTestSetup(t *testing.T) (*reports.Service, *practicums.Service, *sqlcgen.Queries) {
	t.Helper()
	queries := testutil.NewTestQueries(t)
	practicumsService := practicums.NewService(queries)
	return reports.NewService(queries, practicumsService), practicumsService, queries
}

func TestSubmit_Success(t *testing.T) {
	svc, practicumsSvc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	if _, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{}); err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	report, err := svc.Submit(ctx, student.ID, parseDate(t, "2026-01-05"), parseDate(t, "2026-01-11"), "Completed intake training.")
	if err != nil {
		t.Fatalf("Submit returned error: %v", err)
	}
	if report.Status != sqlcgen.ReportStatusSubmitted {
		t.Errorf("Status = %v, want submitted", report.Status)
	}
}

func TestSubmit_DuplicateWeekRejected(t *testing.T) {
	svc, practicumsSvc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	if _, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{}); err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	if _, err := svc.Submit(ctx, student.ID, parseDate(t, "2026-01-05"), parseDate(t, "2026-01-11"), "First submission."); err != nil {
		t.Fatalf("first Submit returned error: %v", err)
	}

	_, err := svc.Submit(ctx, student.ID, parseDate(t, "2026-01-05"), parseDate(t, "2026-01-11"), "Duplicate week.")
	if err == nil {
		t.Fatal("expected an error for a duplicate week_start_date, got nil")
	}
}

func TestSubmit_NoActivePracticum(t *testing.T) {
	svc, _, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	_, err := svc.Submit(ctx, student.ID, parseDate(t, "2026-01-05"), parseDate(t, "2026-01-11"), "No practicum yet.")
	if !errors.Is(err, practicums.ErrNoActivePracticum) {
		t.Fatalf("error = %v, want ErrNoActivePracticum", err)
	}
}

func TestListForStudent_OrderedByMostRecentWeek(t *testing.T) {
	svc, practicumsSvc, queries := newTestSetup(t)
	ctx := context.Background()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	if _, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, parseDate(t, "2026-01-01"), pgtype.Date{}); err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	if _, err := svc.Submit(ctx, student.ID, parseDate(t, "2026-01-05"), parseDate(t, "2026-01-11"), "Week 1"); err != nil {
		t.Fatalf("Submit(week 1) returned error: %v", err)
	}
	if _, err := svc.Submit(ctx, student.ID, parseDate(t, "2026-01-12"), parseDate(t, "2026-01-18"), "Week 2"); err != nil {
		t.Fatalf("Submit(week 2) returned error: %v", err)
	}

	list, err := svc.ListForStudent(ctx, student.ID)
	if err != nil {
		t.Fatalf("ListForStudent returned error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len(list) = %d, want 2", len(list))
	}
	if list[0].Summary != "Week 2" {
		t.Errorf("list[0].Summary = %q, want most recent week first", list[0].Summary)
	}
}
