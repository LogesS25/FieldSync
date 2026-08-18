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

func newTestService(t *testing.T, accessTTL, refreshTTL time.Duration) *auth.Service {
	t.Helper()
	queries := testutil.NewTestQueries(t)
	return auth.NewService(queries, testSecret, accessTTL, refreshTTL)
}

func uniqueEmail(t *testing.T) string {
	t.Helper()
	return fmt.Sprintf("%s-%d@example.com", t.Name(), time.Now().UnixNano())
}

func TestRegister_Success(t *testing.T) {
	svc := newTestService(t, time.Hour, 30*24*time.Hour)
	ctx := context.Background()
	email := uniqueEmail(t)

	session, err := svc.Register(ctx, email, "password123", "Test Student", sqlcgen.UserRoleStudent)
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

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := newTestService(t, time.Hour, 30*24*time.Hour)
	ctx := context.Background()
	email := uniqueEmail(t)

	if _, err := svc.Register(ctx, email, "password123", "First", sqlcgen.UserRoleStudent); err != nil {
		t.Fatalf("first Register returned error: %v", err)
	}

	_, err := svc.Register(ctx, email, "different-password", "Second", sqlcgen.UserRoleFacultySupervisor)
	if !errors.Is(err, auth.ErrEmailTaken) {
		t.Fatalf("Register(duplicate email) error = %v, want ErrEmailTaken", err)
	}
}

func TestLogin_Success(t *testing.T) {
	svc := newTestService(t, time.Hour, 30*24*time.Hour)
	ctx := context.Background()
	email := uniqueEmail(t)

	if _, err := svc.Register(ctx, email, "password123", "Test Student", sqlcgen.UserRoleStudent); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	session, err := svc.Login(ctx, email, "password123")
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if session.User.Email != email {
		t.Errorf("Email = %q, want %q", session.User.Email, email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc := newTestService(t, time.Hour, 30*24*time.Hour)
	ctx := context.Background()
	email := uniqueEmail(t)

	if _, err := svc.Register(ctx, email, "correct-password", "Test Student", sqlcgen.UserRoleStudent); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	_, err := svc.Login(ctx, email, "wrong-password")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("Login(wrong password) error = %v, want ErrInvalidCredentials", err)
	}
}

func TestLogin_NonexistentEmail(t *testing.T) {
	svc := newTestService(t, time.Hour, 30*24*time.Hour)
	ctx := context.Background()

	_, err := svc.Login(ctx, "does-not-exist@example.com", "whatever")
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Fatalf("Login(nonexistent email) error = %v, want ErrInvalidCredentials (must not leak whether the email exists)", err)
	}
}

func TestRefresh_Success(t *testing.T) {
	svc := newTestService(t, time.Hour, 30*24*time.Hour)
	ctx := context.Background()
	email := uniqueEmail(t)

	original, err := svc.Register(ctx, email, "password123", "Test Student", sqlcgen.UserRoleStudent)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	refreshed, err := svc.Refresh(ctx, original.RefreshToken)
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
	svc := newTestService(t, time.Hour, 30*24*time.Hour)
	ctx := context.Background()
	email := uniqueEmail(t)

	original, err := svc.Register(ctx, email, "password123", "Test Student", sqlcgen.UserRoleStudent)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if _, err := svc.Refresh(ctx, original.RefreshToken); err != nil {
		t.Fatalf("first Refresh returned error: %v", err)
	}

	// The original refresh token was revoked by the first Refresh call, so
	// reusing it (e.g. a leaked/stale token) must fail — this is the whole
	// point of rotation.
	_, err = svc.Refresh(ctx, original.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("Refresh(already-rotated token) error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestRefresh_UnknownToken(t *testing.T) {
	svc := newTestService(t, time.Hour, 30*24*time.Hour)
	ctx := context.Background()

	_, err := svc.Refresh(ctx, "a-token-that-was-never-issued")
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("Refresh(unknown token) error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestRefresh_ExpiredToken(t *testing.T) {
	// Negative refresh TTL means the token is already expired the moment
	// it's issued.
	svc := newTestService(t, time.Hour, -time.Minute)
	ctx := context.Background()
	email := uniqueEmail(t)

	session, err := svc.Register(ctx, email, "password123", "Test Student", sqlcgen.UserRoleStudent)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	_, err = svc.Refresh(ctx, session.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("Refresh(expired token) error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestLogout_RevokesRefreshToken(t *testing.T) {
	svc := newTestService(t, time.Hour, 30*24*time.Hour)
	ctx := context.Background()
	email := uniqueEmail(t)

	session, err := svc.Register(ctx, email, "password123", "Test Student", sqlcgen.UserRoleStudent)
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if err := svc.Logout(ctx, session.RefreshToken); err != nil {
		t.Fatalf("Logout returned error: %v", err)
	}

	_, err = svc.Refresh(ctx, session.RefreshToken)
	if !errors.Is(err, auth.ErrInvalidRefreshToken) {
		t.Fatalf("Refresh(logged-out token) error = %v, want ErrInvalidRefreshToken", err)
	}
}

func TestLogout_UnknownTokenIsIdempotent(t *testing.T) {
	svc := newTestService(t, time.Hour, 30*24*time.Hour)
	ctx := context.Background()

	if err := svc.Logout(ctx, "a-token-that-was-never-issued"); err != nil {
		t.Fatalf("Logout(unknown token) returned error %v, want nil (logout must be idempotent)", err)
	}
}
