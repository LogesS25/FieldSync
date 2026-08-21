package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

// CreateTestUser inserts a user with the given role directly (bypassing the
// auth service) so tests for other packages don't need to depend on
// registration semantics.
func CreateTestUser(t *testing.T, queries *sqlcgen.Queries, role sqlcgen.UserRole) sqlcgen.User {
	t.Helper()

	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	email := fmt.Sprintf("%s-%d@example.com", role, time.Now().UnixNano())
	user, err := queries.CreateUser(context.Background(), sqlcgen.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		Role:         role,
		FullName:     "Test " + string(role),
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	return user
}

func CreateTestInstitution(t *testing.T, queries *sqlcgen.Queries) sqlcgen.Institution {
	t.Helper()

	name := fmt.Sprintf("Test Institution %d", time.Now().UnixNano())
	institution, err := queries.CreateInstitution(context.Background(), name)
	if err != nil {
		t.Fatalf("CreateInstitution: %v", err)
	}
	return institution
}

func CreateTestAgency(t *testing.T, queries *sqlcgen.Queries, institutionID pgtype.UUID) sqlcgen.Agency {
	t.Helper()

	name := fmt.Sprintf("Test Agency %d", time.Now().UnixNano())
	agency, err := queries.CreateAgency(context.Background(), sqlcgen.CreateAgencyParams{
		Name:          name,
		InstitutionID: institutionID,
	})
	if err != nil {
		t.Fatalf("CreateAgency: %v", err)
	}
	return agency
}
