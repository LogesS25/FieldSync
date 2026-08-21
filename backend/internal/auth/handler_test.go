package auth_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/testutil"
)

func newTestRouter(t *testing.T) (*gin.Engine, *sqlcgen.Queries) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	svc, queries := newTestService(t, time.Hour, 30*24*time.Hour)
	r := gin.New()
	auth.NewHandler(svc).RegisterRoutes(r)
	return r, queries
}

func doJSON(t *testing.T, r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encoding request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestRegisterHandler_Success(t *testing.T) {
	r, queries := newTestRouter(t)
	institution := testutil.CreateTestInstitution(t, queries)

	rec := doJSON(t, r, http.MethodPost, "/auth/register", map[string]string{
		"email":         fmt.Sprintf("handler-%d@example.com", time.Now().UnixNano()),
		"password":      "password123",
		"fullName":      "Handler Test",
		"role":          "student",
		"institutionId": db.UUIDToString(institution.ID),
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestRegisterHandler_RejectsAdministratorRole(t *testing.T) {
	r, _ := newTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/auth/register", map[string]string{
		"email":    fmt.Sprintf("admin-attempt-%d@example.com", time.Now().UnixNano()),
		"password": "password123",
		"fullName": "Would-Be Admin",
		"role":     "administrator",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (administrator self-registration must be blocked); body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRegisterHandler_RejectsUnknownRole(t *testing.T) {
	r, _ := newTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/auth/register", map[string]string{
		"email":    fmt.Sprintf("unknown-role-%d@example.com", time.Now().UnixNano()),
		"password": "password123",
		"fullName": "Test",
		"role":     "superadmin",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRegisterHandler_RejectsInvalidEmail(t *testing.T) {
	r, queries := newTestRouter(t)
	institution := testutil.CreateTestInstitution(t, queries)

	rec := doJSON(t, r, http.MethodPost, "/auth/register", map[string]string{
		"email":         "not-an-email",
		"password":      "password123",
		"fullName":      "Test",
		"role":          "student",
		"institutionId": db.UUIDToString(institution.ID),
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRegisterHandler_RejectsShortPassword(t *testing.T) {
	r, queries := newTestRouter(t)
	institution := testutil.CreateTestInstitution(t, queries)

	rec := doJSON(t, r, http.MethodPost, "/auth/register", map[string]string{
		"email":         fmt.Sprintf("short-pw-%d@example.com", time.Now().UnixNano()),
		"password":      "short",
		"fullName":      "Test",
		"role":          "student",
		"institutionId": db.UUIDToString(institution.ID),
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRegisterHandler_RejectsMissingFields(t *testing.T) {
	r, _ := newTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/auth/register", map[string]string{
		"email": fmt.Sprintf("missing-fields-%d@example.com", time.Now().UnixNano()),
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRegisterHandler_StudentWithoutInstitutionReturns400(t *testing.T) {
	r, _ := newTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/auth/register", map[string]string{
		"email":    fmt.Sprintf("no-institution-%d@example.com", time.Now().UnixNano()),
		"password": "password123",
		"fullName": "Test",
		"role":     "student",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (student must provide institutionId); body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRegisterHandler_AgencySupervisorWithoutAgencyReturns400(t *testing.T) {
	r, _ := newTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/auth/register", map[string]string{
		"email":    fmt.Sprintf("no-agency-%d@example.com", time.Now().UnixNano()),
		"password": "password123",
		"fullName": "Test",
		"role":     "agency_supervisor",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d (agency supervisor must provide agencyId); body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRegisterHandler_DuplicateEmailReturns409(t *testing.T) {
	r, queries := newTestRouter(t)
	institution := testutil.CreateTestInstitution(t, queries)
	email := fmt.Sprintf("dup-%d@example.com", time.Now().UnixNano())
	payload := map[string]string{
		"email":         email,
		"password":      "password123",
		"fullName":      "Test",
		"role":          "student",
		"institutionId": db.UUIDToString(institution.ID),
	}

	first := doJSON(t, r, http.MethodPost, "/auth/register", payload)
	if first.Code != http.StatusCreated {
		t.Fatalf("first register status = %d, want %d; body = %s", first.Code, http.StatusCreated, first.Body.String())
	}

	second := doJSON(t, r, http.MethodPost, "/auth/register", payload)
	if second.Code != http.StatusConflict {
		t.Fatalf("second register status = %d, want %d; body = %s", second.Code, http.StatusConflict, second.Body.String())
	}
}

func TestLoginHandler_WrongPasswordReturns401(t *testing.T) {
	r, queries := newTestRouter(t)
	institution := testutil.CreateTestInstitution(t, queries)
	email := fmt.Sprintf("login-401-%d@example.com", time.Now().UnixNano())

	reg := doJSON(t, r, http.MethodPost, "/auth/register", map[string]string{
		"email":         email,
		"password":      "correct-password",
		"fullName":      "Test",
		"role":          "student",
		"institutionId": db.UUIDToString(institution.ID),
	})
	if reg.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want %d; body = %s", reg.Code, http.StatusCreated, reg.Body.String())
	}

	rec := doJSON(t, r, http.MethodPost, "/auth/login", map[string]string{
		"email":    email,
		"password": "wrong-password",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestLoginHandler_MalformedJSONReturns400(t *testing.T) {
	r, _ := newTestRouter(t)
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{not valid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRefreshHandler_InvalidTokenReturns401(t *testing.T) {
	r, _ := newTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/auth/refresh", map[string]string{
		"refreshToken": "not-a-real-token",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestLogoutHandler_UnknownTokenReturns204(t *testing.T) {
	r, _ := newTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/auth/logout", map[string]string{
		"refreshToken": "not-a-real-token",
	})
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d (logout must be idempotent even for unknown tokens); body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}
