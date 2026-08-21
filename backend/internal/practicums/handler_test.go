package practicums_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/practicums"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, *sqlcgen.Queries) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	queries := testutil.NewTestQueries(t)
	r := gin.New()
	practicums.NewHandler(practicums.NewService(queries)).RegisterRoutes(r, testSecret)
	return r, queries
}

func tokenFor(t *testing.T, user sqlcgen.User) string {
	t.Helper()
	token, err := auth.GenerateAccessToken(testSecret, db.UUIDToString(user.ID), string(user.Role), time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	return token
}

func doGet(t *testing.T, r *gin.Engine, path, bearerToken string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// Practicum/Placement/SupervisorAssignment creation now happens via
// internal/teamrequests (student-initiated, mutual accept) — see that
// package's handler_test.go for the full happy-path flow exercising
// GET /practicums/me and GET /students end to end. This file only covers
// this package's own role guards and not-found behavior.

func TestGetMyPracticumHandler_RequiresStudentRole(t *testing.T) {
	r, queries := newTestRouter(t)
	supervisor := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)

	rec := doGet(t, r, "/practicums/me", tokenFor(t, supervisor))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestGetMyPracticumHandler_NoActivePracticumReturns404(t *testing.T) {
	r, queries := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	rec := doGet(t, r, "/practicums/me", tokenFor(t, student))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestListStudentsHandler_RequiresSupervisorRole(t *testing.T) {
	r, queries := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	rec := doGet(t, r, "/students", tokenFor(t, student))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}
