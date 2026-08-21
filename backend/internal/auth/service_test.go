package auth_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestService(t *testing.T, accessTTL, refreshTTL time.Duration) (*auth.Service, *sqlcgen.Queries) {
	t.Helper()
	queries := testutil.NewTestQueries(t)
	return auth.NewService(queries, testSecret, accessTTL, refreshTTL), queries
}

func uniqueEmail(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("%s-%d@example.com", t.Name(), time.Now().UnixNano())
}

// registerStudent creates a fresh institution and registers a student
// against it — students require InstitutionID per the business
// requirements (every stakeholder is tied to a university).
func registerStudent(t *testing.T, svc *auth.Service, queries *sqlcgen.Queries, email string) (auth.Session, error) {
	t.Helper()
	institution := testutil.CreateTestInstitution(t, queries)
	return svc.Register(context.Background(), auth.RegisterInput{
		Email:         email,
		Password:      "password123",
		FullName:      "Test Student",
		Role:          sqlcgen.UserRoleStudent,
		InstitutionID: institution.ID,
	})
}

func TestRegister_Success(t *testing.T) {
	svc, queries := newTestService(t, time.Hour, 30*24*time.Hour)
	email := uniqueEmail(t)

	session, err := registerStudent(t, svc, queries, email)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if session.User.Email != email {
		t.Errorf("Email = %q, want %q", session.User.Email, email)
	}
	if session.User.Role != sqlcgen.UserRoleStudent {
		t.Errorf("Role = %q, want %q", session.User.Role, sqlcgen.UserRoleStudent)
	}
	if session.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if session.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestRegister_StudentRequiresInstitution(t *testing.T) {
	svc, _ := newTestService(t, time.Hour, 30*24*time.Hour)

	_, err := svc.Register(context.Background(), auth.RegisterInput{
		Email:    uniqueEmail(t),
		Password: "password123",
		FullName: "Test Student",
		Role:     sqlcgen.UserRoleStudent,
	})
	if !errors.Is(err, auth.ErrInstitutionRequired) {
		t.Fatalf("error = %v, want ErrInstitutionRequired", err)
	}
}

func TestRegister_AgencySupervisorRequiresAgency(t *testing.T) {
	svc, _ := newTestService(t, time.Hour, 30*24*time.Hour)

	_, err := svc.Register(context.Background(), auth.RegisterInput{
		Email:    uniqueEmail(t),
		Password: "password123",
		FullName: "Test Agency Supervisor",
		Role:     sqlcgen.UserRoleAgencySupervisor,
	})
	if !errors.Is(err, auth.ErrAgencyRequired) {
		t.Fatalf("error = %v, want ErrAgencyRequired", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, queries := newTestService(t, time.Hour, 30*24*time.Hour)
	email := uniqueEmail(t)

	if _, err := registerStudent(t, svc, queries, email); err != nil {
		t.Fatalf("first Register returned error: %v", err)
	}

	institution := testutil.CreateTestInstitution(t, queries)
	_, err := svc.Register(context.Background(), auth.RegisterInput{
		Email:         email,
		Password:      "different-password",
		FullName:      "Second",
		Role:          sqlcgen.UserRoleFacultySupervisor,
		InstitutionID: institution.ID,
	})
	if !errors.Is(err, auth.ErrEmailTaken) {
		t.Fatalf("Register(duplicate email) error = %v, want ErrEmailTaken", err)
	}
}

func TestLogin_Success(t *testing.T) {
	svc, queries := newTestService(t, time.Hour, 30*24*time.Hour)
	email := uniqueEmail(t)

	if _, err := registerStudent(t, svc, queries, email); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	session, err := svc.Login(context.Background(), email, "password123")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if session.User.Email != email {
		t.Errorf("Email = %q, want %q", session.User.Email, email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, queries := newTestService(t, time.Hour, 30*24*time.Hour)
	email := uniqueEmail(t)

	if _, err := registerStudent(t, svc, queries, email); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	_, err := svc.Login(context.Background(), email, "wrong-password")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("Login(wrong password) error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_NonexistentEmail(t *testing.T) {
	svc, _ := newTestService(t, time.Hour, 30*24*time.Hour)

	_, err := svc.Login(context.Background(), "does-not-exist@example.com", "whatever")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("Login(nonexistent email) error = %v, want ErrInvalidCredentials (must not leak whether the email exists)", err)
	}
}

func TestRefresh_Success(t *testing.T) {
	svc, queries := newTestService(t, time.Hour, 30*24*time.Hour)
	email := uniqueEmail(t)

	original, err := registerStudent(t, svc, queries, email)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	refreshed, err := svc.Refresh(context.Background(), original.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}

	if refreshed.RefreshToken == original.RefreshToken {
		t.Error("expected refresh to rotate to a new refresh token, got the same one back")
	}
	if refreshed.User.Email != email {
		t.Errorf("Email = %q, want %q", refreshed.User.Email, email)
	}
}

func TestRefresh_RotatedTokenCannotBeReused(t *testing.T) {
	svc, queries := newTestService(t, time.Hour, 30*24*time.Hour)
	email := uniqueEmail(t)

	original, err := registerStudent(t, svc, queries, email)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if _, err := svc.Refresh(context.Background(), original.RefreshToken); err != nil {
		t.Fatalf("first Refresh returned error: %v", err)
	}

	// The original refresh token was revoked by the first Refresh call, so
	// reusing it (e.g. a leaked/stale token) must fail — this is the whole
	// point of rotation.
	_, err = svc.Refresh(context.Background(), original.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("Refresh(already-rotated token) error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestRefresh_UnknownToken(t *testing.T) {
	svc, _ := newTestService(t, time.Hour, 30*24*time.Hour)

	_, err := svc.Refresh(context.Background(), "a-token-that-was-never-issued")
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("Refresh(unknown token) error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestRefresh_ExpiredToken(t *testing.T) {
	// Negative refresh TTL means the token is already expired the moment
	// it's issued.
	svc, queries := newTestService(t, time.Hour, -time.Minute)
	email := uniqueEmail(t)

	session, err := registerStudent(t, svc, queries, email)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	_, err = svc.Refresh(context.Background(), session.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("Refresh(expired token) error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestLogout_RevokesRefreshToken(t *testing.T) {
	svc, queries := newTestService(t, time.Hour, 30*24*time.Hour)
	email := uniqueEmail(t)

	session, err := registerStudent(t, svc, queries, email)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if err := svc.Logout(context.Background(), session.RefreshToken); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}

	_, err = svc.Refresh(context.Background(), session.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("Refresh(logged-out token) error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestLogout_UnknownTokenIsIdempotent(t *testing.T) {
	svc, _ := newTestService(t, time.Hour, 30*24*time.Hour)

	if err := svc.Logout(context.Background(), "a-token-that-was-never-issued"); err != nil {
		t.Fatalf("Logout(unknown token) returned error %v, want nil (logout must be idempotent)", err)
	}
}
