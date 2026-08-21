package fieldworkcomponents_test

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
	"github.com/fieldsync/backend/internal/fieldworkcomponents"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, *sqlcgen.Queries) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	queries := testutil.NewTestQueries(t)
	r := gin.New()
	fieldworkcomponents.NewHandler(queries).RegisterRoutes(r, testSecret)
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

func TestCreate_RequiresAdmin(t *testing.T) {
	r, queries := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)

	rec := doJSON(t, r, http.MethodPost, "/fieldwork-components", tokenFor(t, student), map[string]string{
		"name":          "Casework",
		"institutionId": db.UUIDToString(institution.ID),
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCreate_Success(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institution := testutil.CreateTestInstitution(t, queries)

	rec := doJSON(t, r, http.MethodPost, "/fieldwork-components", tokenFor(t, admin), map[string]string{
		"name":          fmt.Sprintf("Casework %d", time.Now().UnixNano()),
		"institutionId": db.UUIDToString(institution.ID),
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
}

func TestUpdateAndDelete_UniversityControlOverOwnList(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institution := testutil.CreateTestInstitution(t, queries)
	adminToken := tokenFor(t, admin)

	createRec := doJSON(t, r, http.MethodPost, "/fieldwork-components", adminToken, map[string]string{
		"name":          fmt.Sprintf("Casework %d", time.Now().UnixNano()),
		"institutionId": db.UUIDToString(institution.ID),
	})
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}
	id := created["id"].(string)

	newName := fmt.Sprintf("Group Work %d", time.Now().UnixNano())
	updateRec := doJSON(t, r, http.MethodPatch, "/fieldwork-components/"+id, adminToken, map[string]string{"name": newName})
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d; body = %s", updateRec.Code, http.StatusOK, updateRec.Body.String())
	}
	var updated map[string]any
	if err := json.Unmarshal(updateRec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decoding update response: %v", err)
	}
	if updated["name"] != newName {
		t.Errorf("name = %v, want %v", updated["name"], newName)
	}

	deleteRec := doJSON(t, r, http.MethodDelete, "/fieldwork-components/"+id, adminToken, nil)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d; body = %s", deleteRec.Code, http.StatusNoContent, deleteRec.Body.String())
	}

	listRec := doJSON(t, r, http.MethodGet, "/fieldwork-components", adminToken, nil)
	var list []map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decoding list response: %v", err)
	}
	for _, item := range list {
		if item["id"] == id {
			t.Fatalf("deleted component still present in list: %+v", item)
		}
	}
}

func TestListMine_ScopedToOwnInstitution(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institutionA := testutil.CreateTestInstitution(t, queries)
	institutionB := testutil.CreateTestInstitution(t, queries)
	adminToken := tokenFor(t, admin)

	nameA := fmt.Sprintf("Component A %d", time.Now().UnixNano())
	nameB := fmt.Sprintf("Component B %d", time.Now().UnixNano())
	doJSON(t, r, http.MethodPost, "/fieldwork-components", adminToken, map[string]string{"name": nameA, "institutionId": db.UUIDToString(institutionA.ID)})
	doJSON(t, r, http.MethodPost, "/fieldwork-components", adminToken, map[string]string{"name": nameB, "institutionId": db.UUIDToString(institutionB.ID)})

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

	rec := doJSON(t, r, http.MethodGet, "/fieldwork-components/mine", tokenFor(t, student), nil)
	var results []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(results) != 1 || results[0]["name"] != nameA {
		t.Fatalf("results = %+v, want exactly institution A's component", results)
	}
}

func TestCreate_DuplicateNameWithinUniversityReturns409(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institution := testutil.CreateTestInstitution(t, queries)
	adminToken := tokenFor(t, admin)
	name := fmt.Sprintf("Casework %d", time.Now().UnixNano())
	body := map[string]string{"name": name, "institutionId": db.UUIDToString(institution.ID)}

	first := doJSON(t, r, http.MethodPost, "/fieldwork-components", adminToken, body)
	if first.Code != http.StatusCreated {
		t.Fatalf("first create status = %d, want %d; body = %s", first.Code, http.StatusCreated, first.Body.String())
	}
	second := doJSON(t, r, http.MethodPost, "/fieldwork-components", adminToken, body)
	if second.Code != http.StatusConflict {
		t.Fatalf("second create status = %d, want %d; body = %s", second.Code, http.StatusConflict, second.Body.String())
	}
}
