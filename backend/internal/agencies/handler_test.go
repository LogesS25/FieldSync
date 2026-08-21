package agencies_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/agencies"
	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, *sqlcgen.Queries) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	queries := testutil.NewTestQueries(t)
	r := gin.New()
	agencies.NewHandler(queries).RegisterRoutes(r, testSecret)
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

func TestCreateAgency_RequiresAdmin(t *testing.T) {
	r, queries := newTestRouter(t)
	supervisor := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAgencySupervisor)
	institution := testutil.CreateTestInstitution(t, queries)

	rec := doJSON(t, r, http.MethodPost, "/agencies", tokenFor(t, supervisor), map[string]string{
		"name":          "Test Agency",
		"institutionId": db.UUIDToString(institution.ID),
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCreateAgency_Success(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institution := testutil.CreateTestInstitution(t, queries)

	rec := doJSON(t, r, http.MethodPost, "/agencies", tokenFor(t, admin), map[string]string{
		"name":          fmt.Sprintf("Test Agency %d", time.Now().UnixNano()),
		"institutionId": db.UUIDToString(institution.ID),
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestCreateAgency_InvalidInstitutionReturns400(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)

	rec := doJSON(t, r, http.MethodPost, "/agencies", tokenFor(t, admin), map[string]string{
		"name":          fmt.Sprintf("Test Agency %d", time.Now().UnixNano()),
		"institutionId": "00000000-0000-0000-0000-000000000000",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCreateAgency_DuplicateNameReturns409(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institution := testutil.CreateTestInstitution(t, queries)
	name := fmt.Sprintf("Duplicate Agency %d", time.Now().UnixNano())
	body := map[string]string{"name": name, "institutionId": db.UUIDToString(institution.ID)}

	first := doJSON(t, r, http.MethodPost, "/agencies", tokenFor(t, admin), body)
	if first.Code != http.StatusCreated {
		t.Fatalf("first create status = %d, want %d; body = %s", first.Code, http.StatusCreated, first.Body.String())
	}

	second := doJSON(t, r, http.MethodPost, "/agencies", tokenFor(t, admin), body)
	if second.Code != http.StatusConflict {
		t.Fatalf("second create status = %d, want %d; body = %s", second.Code, http.StatusConflict, second.Body.String())
	}
}

func TestListAgencies_RequiresAdmin(t *testing.T) {
	r, queries := newTestRouter(t)
	supervisor := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAgencySupervisor)

	rec := doJSON(t, r, http.MethodGet, "/agencies", tokenFor(t, supervisor), nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestListMine_ScopedToOwnInstitution(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institutionA := testutil.CreateTestInstitution(t, queries)
	institutionB := testutil.CreateTestInstitution(t, queries)

	nameA := fmt.Sprintf("Agency A %d", time.Now().UnixNano())
	nameB := fmt.Sprintf("Agency B %d", time.Now().UnixNano())
	createA := doJSON(t, r, http.MethodPost, "/agencies", tokenFor(t, admin), map[string]string{"name": nameA, "institutionId": db.UUIDToString(institutionA.ID)})
	if createA.Code != http.StatusCreated {
		t.Fatalf("create A status = %d, want %d; body = %s", createA.Code, http.StatusCreated, createA.Body.String())
	}
	createB := doJSON(t, r, http.MethodPost, "/agencies", tokenFor(t, admin), map[string]string{"name": nameB, "institutionId": db.UUIDToString(institutionB.ID)})
	if createB.Code != http.StatusCreated {
		t.Fatalf("create B status = %d, want %d; body = %s", createB.Code, http.StatusCreated, createB.Body.String())
	}

	student, err := queries.CreateUser(t.Context(), sqlcgen.CreateUserParams{
		Email:         fmt.Sprintf("student-%d@example.com", time.Now().UnixNano()),
		PasswordHash:  "irrelevant",
		Role:          sqlcgen.UserRoleStudent,
		FullName:      "Test Student",
		InstitutionID: institutionA.ID,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	rec := doJSON(t, r, http.MethodGet, "/agencies/mine", tokenFor(t, student), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var results []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1 (must not see other university's agencies)", len(results))
	}
	if results[0]["name"] != nameA {
		t.Errorf("name = %v, want %v", results[0]["name"], nameA)
	}
}
