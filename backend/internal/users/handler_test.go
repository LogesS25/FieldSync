package users_test

import (
	"context"
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
	"github.com/fieldsync/backend/internal/users"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	queries := testutil.NewTestQueries(t)
	r := gin.New()
	users.NewHandler(queries).RegisterRoutes(r, testSecret)
	return r
}

func TestMeHandler_ReturnsAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	queries := testutil.NewTestQueries(t)
	r := gin.New()
	users.NewHandler(queries).RegisterRoutes(r, testSecret)

	email := fmt.Sprintf("me-%d@example.com", time.Now().UnixNano())
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	user, err := queries.CreateUser(context.Background(), sqlcgen.CreateUserParams{
		Email:        email,
		PasswordHash: hash,
		Role:         sqlcgen.UserRoleStudent,
		FullName:     "Me Test",
	})
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	token, err := auth.GenerateAccessToken(testSecret, db.UUIDToString(user.ID), string(user.Role), time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decoding response body: %v", err)
	}
	if body["email"] != email {
		t.Errorf("email = %v, want %v", body["email"], email)
	}
}

func TestMeHandler_NoToken(t *testing.T) {
	r := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMeHandler_TokenForDeletedUser(t *testing.T) {
	r := newTestRouter(t)

	// A well-formed, validly-signed token for a user ID that doesn't exist
	// in the DB (e.g. the user was deleted after the token was issued).
	token, err := auth.GenerateAccessToken(testSecret, "00000000-0000-0000-0000-000000000000", "student", time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
