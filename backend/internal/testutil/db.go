// Package testutil provides shared test infrastructure. It is only ever
// imported from _test.go files.
package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

const defaultTestDatabaseURL = "postgres://fieldsync:fieldsync@localhost:5433/fieldsync?sslmode=disable"

// NewTestQueries opens a connection to the local dev Postgres, begins a
// transaction, and returns sqlc queries scoped to that transaction. The
// transaction is always rolled back in cleanup, so tests can freely
// register/login/etc. without leaving data behind or needing a separate
// test database.
//
// Requires `docker compose up -d postgres` (see repo README) with
// migrations applied. Skips the test if the database is unreachable, so
// `go test ./...` doesn't hard-fail for someone who hasn't started it.
func NewTestQueries(t *testing.T) *sqlcgen.Queries {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = defaultTestDatabaseURL
	}

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		t.Skipf("skipping: could not connect to test database at %s: %v", dbURL, err)
	}
	t.Cleanup(func() { _ = conn.Close(ctx) })

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("beginning test transaction: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	return sqlcgen.New(tx)
}
