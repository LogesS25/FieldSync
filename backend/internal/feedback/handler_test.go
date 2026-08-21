package feedback_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/feedback"
	"github.com/fieldsync/backend/internal/practicums"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

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

func TestSubmitHandler_RequiresSupervisorRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	queries := testutil.NewTestQueries(t)
	r := gin.New()
	feedback.NewHandler(feedback.NewService(queries)).RegisterRoutes(r, testSecret)

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	fakeID, err := db.ParseUUID("00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("ParseUUID: %v", err)
	}

	rec := doJSON(t, r, http.MethodPost, "/feedback", tokenFor(t, student), map[string]string{
		"practicumId":   db.UUIDToString(fakeID),
		"weekStartDate": "2026-01-04",
		"feedback":      "Should be rejected.",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestFullFeedbackFlow_SubmitThenStudentSees(t *testing.T) {
	gin.SetMode(gin.TestMode)
	queries := testutil.NewTestQueries(t)
	practicumsSvc := practicums.NewService(queries)
	r := gin.New()
	feedback.NewHandler(feedback.NewService(queries)).RegisterRoutes(r, testSecret)
	ctx := t.Context()

	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)
	startDate, err := db.ParseDate("2026-01-01")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	practicum, err := practicumsSvc.CreatePracticum(ctx, student.ID, institution.ID, startDate, pgtype.Date{})
	if err != nil {
		t.Fatalf("CreatePracticum: %v", err)
	}
	faculty := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)
	if _, err := practicumsSvc.CreateSupervisorAssignment(ctx, practicum.ID, faculty.ID); err != nil {
		t.Fatalf("CreateSupervisorAssignment: %v", err)
	}

	submitRec := doJSON(t, r, http.MethodPost, "/feedback", tokenFor(t, faculty), map[string]string{
		"practicumId":   db.UUIDToString(practicum.ID),
		"weekStartDate": "2026-01-04",
		"feedback":      "Good communication with the agency this week.",
	})
	if submitRec.Code != http.StatusCreated {
		t.Fatalf("submit status = %d, want %d; body = %s", submitRec.Code, http.StatusCreated, submitRec.Body.String())
	}

	studentRec := doJSON(t, r, http.MethodGet, "/feedback", tokenFor(t, student), nil)
	var entries []map[string]any
	if err := json.Unmarshal(studentRec.Body.Bytes(), &entries); err != nil {
		t.Fatalf("decoding student feedback response: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}
}
