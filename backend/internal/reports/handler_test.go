package reports_test

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
	"github.com/fieldsync/backend/internal/practicums"
	"github.com/fieldsync/backend/internal/reports"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, *sqlcgen.Queries, *practicums.Service) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	queries := testutil.NewTestQueries(t)
	practicumsService := practicums.NewService(queries)
	r := gin.New()
	reports.NewHandler(reports.NewService(queries, practicumsService)).RegisterRoutes(r, testSecret)
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

func TestSubmitWeeklyReportHandler_RejectsMissingSummary(t *testing.T) {
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

	rec := doJSON(t, r, http.MethodPost, "/weekly-reports", tokenFor(t, student), map[string]string{
		"weekStartDate": "2026-01-05",
		"weekEndDate":   "2026-01-11",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestSubmitWeeklyReportHandler_DuplicateWeekReturns409(t *testing.T) {
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

	payload := map[string]string{
		"weekStartDate": "2026-01-05",
		"weekEndDate":   "2026-01-11",
		"summary":       "Week one summary.",
	}
	first := doJSON(t, r, http.MethodPost, "/weekly-reports", token, payload)
	if first.Code != http.StatusCreated {
		t.Fatalf("first status = %d, want %d; body = %s", first.Code, http.StatusCreated, first.Body.String())
	}

	second := doJSON(t, r, http.MethodPost, "/weekly-reports", token, payload)
	if second.Code != http.StatusConflict {
		t.Fatalf("second status = %d, want %d; body = %s", second.Code, http.StatusConflict, second.Body.String())
	}
}

func TestWeeklyReportFlow_SubmitThenList(t *testing.T) {
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

	submitRec := doJSON(t, r, http.MethodPost, "/weekly-reports", token, map[string]string{
		"weekStartDate": "2026-01-05",
		"weekEndDate":   "2026-01-11",
		"summary":       "Worked on case documentation.",
	})
	if submitRec.Code != http.StatusCreated {
		t.Fatalf("submit status = %d, want %d; body = %s", submitRec.Code, http.StatusCreated, submitRec.Body.String())
	}

	listRec := doJSON(t, r, http.MethodGet, "/weekly-reports", token, nil)
	var reportList []map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &reportList); err != nil {
		t.Fatalf("decoding list response: %v", err)
	}
	if len(reportList) != 1 {
		t.Fatalf("len(reportList) = %d, want 1", len(reportList))
	}
	if reportList[0]["status"] != "submitted" {
		t.Errorf("status = %v, want submitted", reportList[0]["status"])
	}
}
