package attendance_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/attendance"
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
	attendance.NewHandler(attendance.NewService(queries, practicumsService)).RegisterRoutes(r, testSecret)
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

func TestCreateAttendanceHandler_RejectsZeroHours(t *testing.T) {
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

	rec := doJSON(t, r, http.MethodPost, "/attendance", tokenFor(t, student), map[string]any{
		"attendanceDate": "2026-01-05",
		"hours":          0,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCreateAttendanceHandler_RejectsOver24Hours(t *testing.T) {
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

	rec := doJSON(t, r, http.MethodPost, "/attendance", tokenFor(t, student), map[string]any{
		"attendanceDate": "2026-01-05",
		"hours":          25,
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestAttendanceFlow_CreateListSummary(t *testing.T) {
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

	create1 := doJSON(t, r, http.MethodPost, "/attendance", token, map[string]any{"attendanceDate": "2026-01-05", "hours": 6})
	if create1.Code != http.StatusCreated {
		t.Fatalf("create1 status = %d, want %d; body = %s", create1.Code, http.StatusCreated, create1.Body.String())
	}
	create2 := doJSON(t, r, http.MethodPost, "/attendance", token, map[string]any{"attendanceDate": "2026-01-06", "hours": 4})
	if create2.Code != http.StatusCreated {
		t.Fatalf("create2 status = %d, want %d; body = %s", create2.Code, http.StatusCreated, create2.Body.String())
	}

	listRec := doJSON(t, r, http.MethodGet, "/attendance", token, nil)
	var records []map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &records); err != nil {
		t.Fatalf("decoding list response: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("len(records) = %d, want 2", len(records))
	}

	summaryRec := doJSON(t, r, http.MethodGet, "/attendance/summary", token, nil)
	if summaryRec.Code != http.StatusOK {
		t.Fatalf("summary status = %d, want %d; body = %s", summaryRec.Code, http.StatusOK, summaryRec.Body.String())
	}
	var summary map[string]any
	if err := json.Unmarshal(summaryRec.Body.Bytes(), &summary); err != nil {
		t.Fatalf("decoding summary response: %v", err)
	}
	if summary["totalHours"] != float64(10) {
		t.Errorf("totalHours = %v, want 10", summary["totalHours"])
	}
}
