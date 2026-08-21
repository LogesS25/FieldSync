package notifications_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/notifications"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, *sqlcgen.Queries, *notifications.Service) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	queries := testutil.NewTestQueries(t)
	svc := notifications.NewService(queries)
	r := gin.New()
	notifications.NewHandler(svc).RegisterRoutes(r, testSecret)
	return r, queries, svc
}

func tokenFor(t *testing.T, user sqlcgen.User) string {
	t.Helper()
	token, err := auth.GenerateAccessToken(testSecret, db.UUIDToString(user.ID), string(user.Role), time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}
	return token
}

func doJSON(t *testing.T, r *gin.Engine, method, path, bearerToken string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, &bytes.Buffer{})
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestListHandler_ReturnsOwnNotifications(t *testing.T) {
	r, queries, svc := newTestRouter(t)
	user := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	if _, err := svc.Create(t.Context(), user.ID, "hello"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	rec := doJSON(t, r, http.MethodGet, "/notifications", tokenFor(t, user))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var list []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(list) != 1 || list[0]["message"] != "hello" {
		t.Fatalf("list = %+v, want one notification with message 'hello'", list)
	}
}

func TestMarkReadHandler_RejectsOtherUser(t *testing.T) {
	r, queries, svc := newTestRouter(t)
	user := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	other := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	n, err := svc.Create(t.Context(), user.ID, "hello")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	rec := doJSON(t, r, http.MethodPost, "/notifications/"+db.UUIDToString(n.ID)+"/read", tokenFor(t, other))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestMarkAllReadHandler(t *testing.T) {
	r, queries, svc := newTestRouter(t)
	user := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	if _, err := svc.Create(t.Context(), user.ID, "a"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	rec := doJSON(t, r, http.MethodPost, "/notifications/read-all", tokenFor(t, user))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
}
