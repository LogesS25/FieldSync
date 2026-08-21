package attendance_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/attendance"
	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, testFixture) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	f := newFixture(t)
	r := gin.New()
	attendance.NewHandler(f.svc).RegisterRoutes(r, testSecret)
	return r, f
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

func TestCreateHandler_RejectsInvalidSession(t *testing.T) {
	r, f := newTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/attendance", tokenFor(t, f.student), map[string]any{
		"attendanceDate": "2026-01-05",
		"session":        "afternoon",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestAgencyReviewHandler_RequiresAgencyRole(t *testing.T) {
	r, f := newTestRouter(t)

	createRec := doJSON(t, r, http.MethodPost, "/attendance", tokenFor(t, f.student), map[string]any{
		"attendanceDate": "2026-01-05",
		"session":        "morning",
	})
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}

	rec := doJSON(t, r, http.MethodPost, "/attendance/"+created["id"].(string)+"/agency-review", tokenFor(t, f.faculty), map[string]string{
		"decision": "approved",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (only agency_supervisor role may agency-review); body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestFullApprovalFlow_AgencyThenFaculty(t *testing.T) {
	r, f := newTestRouter(t)
	studentToken := tokenFor(t, f.student)

	createRec := doJSON(t, r, http.MethodPost, "/attendance", studentToken, map[string]any{
		"attendanceDate": "2026-01-05",
		"session":        "morning",
		"hours":          6,
	})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}
	recordID := created["id"].(string)

	facultyTooEarly := doJSON(t, r, http.MethodPost, "/attendance/"+recordID+"/faculty-review", tokenFor(t, f.faculty), map[string]string{"decision": "approved"})
	if facultyTooEarly.Code != http.StatusConflict {
		t.Fatalf("premature faculty review status = %d, want %d; body = %s", facultyTooEarly.Code, http.StatusConflict, facultyTooEarly.Body.String())
	}

	agencyRec := doJSON(t, r, http.MethodPost, "/attendance/"+recordID+"/agency-review", tokenFor(t, f.agencySup), map[string]string{"decision": "approved"})
	if agencyRec.Code != http.StatusOK {
		t.Fatalf("agency review status = %d, want %d; body = %s", agencyRec.Code, http.StatusOK, agencyRec.Body.String())
	}

	facultyRec := doJSON(t, r, http.MethodPost, "/attendance/"+recordID+"/faculty-review", tokenFor(t, f.faculty), map[string]string{"decision": "approved"})
	if facultyRec.Code != http.StatusOK {
		t.Fatalf("faculty review status = %d, want %d; body = %s", facultyRec.Code, http.StatusOK, facultyRec.Body.String())
	}

	listRec := doJSON(t, r, http.MethodGet, "/attendance", studentToken, nil)
	var records []map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &records); err != nil {
		t.Fatalf("decoding list response: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len(records) = %d, want 1", len(records))
	}
	if records[0]["agencyStatus"] != "approved" || records[0]["facultyStatus"] != "approved" {
		t.Errorf("record = %+v, want both statuses approved", records[0])
	}
}
