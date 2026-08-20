package activities_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/activities"
	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/practicums"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, *sqlcgen.Queries, *practicums.Service) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	queries := testutil.NewTestQueries(t)
	practicumsService := practicums.NewService(queries)
	r := gin.New()
	activities.NewHandler(activities.NewService(queries, practicumsService)).RegisterRoutes(r, testSecret)
	return r, queries, practicumsService
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

func TestCreateFieldActivityHandler_RequiresStudentRole(t *testing.T) {
	r, queries, _ := newTestRouter(t)
	supervisor := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)

	rec := doJSON(t, r, http.MethodPost, "/field-activities", tokenFor(t, supervisor), map[string]string{
		"activityDate": "2026-01-05",
		"description":  "Should be rejected",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCreateFieldActivityHandler_NoActivePracticumReturns409(t *testing.T) {
	r, queries, _ := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	rec := doJSON(t, r, http.MethodPost, "/field-activities", tokenFor(t, student), map[string]string{
		"activityDate": "2026-01-05",
		"description":  "No practicum yet",
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestFieldActivityFlow_CreateThenList(t *testing.T) {
	r, queries, practicumsSvc := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)

	startDate, err := db.ParseDate("2026-01-01")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	if _, err := practicumsSvc.CreatePracticum(t.Context(), student.ID, institution.ID, startDate, pgtype.Date{}); err != nil {
		t.Fatalf("CreatePracticum returned error: %v", err)
	}

	token := tokenFor(t, student)
	createRec := doJSON(t, r, http.MethodPost, "/field-activities", token, map[string]string{
		"activityDate": "2026-01-05",
		"description":  "Attended team meeting.",
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}

	listRec := doJSON(t, r, http.MethodGet, "/field-activities", token, nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d; body = %s", listRec.Code, http.StatusOK, listRec.Body.String())
	}

	var results []map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &results); err != nil {
		t.Fatalf("decoding list response: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0]["description"] != "Attended team meeting." {
		t.Errorf("description = %v", results[0]["description"])
	}
}
