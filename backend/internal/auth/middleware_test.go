package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/auth"
)

func newProtectedRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", auth.RequireAuth(secret), func(c *gin.Context) {
		userID, _ := c.Get(auth.ContextUserIDKey)
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})
	r.GET("/student-only", auth.RequireAuth(secret), auth.RequireRole("student"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func doGet(r *gin.Engine, path, bearerToken string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	r := newProtectedRouter(testSecret)

	rec := doGet(r, "/protected", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_MalformedHeader(t *testing.T) {
	r := newProtectedRouter(testSecret)
	gin.SetMode(gin.TestMode)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	r := newProtectedRouter(testSecret)

	token, err := auth.GenerateAccessToken(testSecret, "user-abc", "student", time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	rec := doGet(r, "/protected", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	r := newProtectedRouter(testSecret)

	token, err := auth.GenerateAccessToken(testSecret, "user-abc", "student", -time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	rec := doGet(r, "/protected", token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_WrongSecret(t *testing.T) {
	r := newProtectedRouter(testSecret)

	token, err := auth.GenerateAccessToken("a-different-secret", "user-abc", "student", time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	rec := doGet(r, "/protected", token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireRole_AllowsMatchingRole(t *testing.T) {
	r := newProtectedRouter(testSecret)

	token, err := auth.GenerateAccessToken(testSecret, "user-abc", "student", time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	rec := doGet(r, "/student-only", token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRequireRole_BlocksNonMatchingRole(t *testing.T) {
	r := newProtectedRouter(testSecret)

	token, err := auth.GenerateAccessToken(testSecret, "user-abc", "agency_supervisor", time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	rec := doGet(r, "/student-only", token)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}
