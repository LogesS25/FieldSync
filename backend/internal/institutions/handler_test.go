package institutions_test

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
	"github.com/fieldsync/backend/internal/institutions"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, *sqlcgen.Queries) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	queries := testutil.NewTestQueries(t)
	r := gin.New()
	institutions.NewHandler(queries).RegisterRoutes(r, testSecret)
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

func doJSON(t *testing.T, r *gin.Engine, method, path, bearerToken string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encoding request body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCreateInstitution_RequiresAdmin(t *testing.T) {
	r, queries := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	rec := doJSON(t, r, http.MethodPost, "/institutions", tokenFor(t, student), map[string]string{"name": "Test University"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCreateInstitution_Unauthenticated(t *testing.T) {
	r, _ := newTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/institutions", "", map[string]string{"name": "Test University"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestCreateInstitution_Success(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)

	rec := doJSON(t, r, http.MethodPost, "/institutions", tokenFor(t, admin), map[string]string{
		"name": fmt.Sprintf("Test University %d", time.Now().UnixNano()),
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestCreateInstitution_DuplicateNameReturns409(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	name := fmt.Sprintf("Duplicate University %d", time.Now().UnixNano())

	first := doJSON(t, r, http.MethodPost, "/institutions", tokenFor(t, admin), map[string]string{"name": name})
	if first.Code != http.StatusCreated {
		t.Fatalf("first create status = %d, want %d; body = %s", first.Code, http.StatusCreated, first.Body.String())
	}

	second := doJSON(t, r, http.MethodPost, "/institutions", tokenFor(t, admin), map[string]string{"name": name})
	if second.Code != http.StatusConflict {
		t.Fatalf("second create status = %d, want %d; body = %s", second.Code, http.StatusConflict, second.Body.String())
	}
}

func TestListInstitutions_RequiresAdmin(t *testing.T) {
	r, queries := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	rec := doJSON(t, r, http.MethodGet, "/institutions", tokenFor(t, student), nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestListInstitutions_IncludesCreated(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	name := fmt.Sprintf("Listed University %d", time.Now().UnixNano())

	created := doJSON(t, r, http.MethodPost, "/institutions", tokenFor(t, admin), map[string]string{"name": name})
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", created.Code, http.StatusCreated, created.Body.String())
	}

	listRec := doJSON(t, r, http.MethodGet, "/institutions", tokenFor(t, admin), nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d; body = %s", listRec.Code, http.StatusOK, listRec.Body.String())
	}

	var results []map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &results); err != nil {
		t.Fatalf("decoding list response: %v", err)
	}

	found := false
	for _, r := range results {
		if r["name"] == name {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected created institution %q to appear in list", name)
	}
}
