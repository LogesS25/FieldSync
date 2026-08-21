package reports_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/reports"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, testFixture) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	f := newFixture(t)
	r := gin.New()
	reports.NewHandler(f.svc).RegisterRoutes(r, testSecret)
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

func TestSubmitHandler_RequiresStudentRole(t *testing.T) {
	r, f := newTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/consolidated-reports", tokenFor(t, f.faculty), map[string]string{
		"summary": "Should be rejected.",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestGetMineHandler_NotFoundReturns404(t *testing.T) {
	r, f := newTestRouter(t)

	rec := doJSON(t, r, http.MethodGet, "/consolidated-reports/me", tokenFor(t, f.student), nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestFullReviewFlow_AgencyThenFaculty(t *testing.T) {
	r, f := newTestRouter(t)
	studentToken := tokenFor(t, f.student)

	submitRec := doJSON(t, r, http.MethodPost, "/consolidated-reports", studentToken, map[string]string{
		"summary": "Full fieldwork summary.",
	})
	if submitRec.Code != http.StatusCreated {
		t.Fatalf("submit status = %d, want %d; body = %s", submitRec.Code, http.StatusCreated, submitRec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(submitRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding submit response: %v", err)
	}
	reportID := created["id"].(string)

	facultyTooEarly := doJSON(t, r, http.MethodPost, "/consolidated-reports/"+reportID+"/faculty-review", tokenFor(t, f.faculty), map[string]string{"decision": "approved"})
	if facultyTooEarly.Code != http.StatusConflict {
		t.Fatalf("premature faculty review status = %d, want %d; body = %s", facultyTooEarly.Code, http.StatusConflict, facultyTooEarly.Body.String())
	}

	agencyRec := doJSON(t, r, http.MethodPost, "/consolidated-reports/"+reportID+"/agency-review", tokenFor(t, f.agencySup), map[string]string{"decision": "approved"})
	if agencyRec.Code != http.StatusOK {
		t.Fatalf("agency review status = %d, want %d; body = %s", agencyRec.Code, http.StatusOK, agencyRec.Body.String())
	}

	facultyRec := doJSON(t, r, http.MethodPost, "/consolidated-reports/"+reportID+"/faculty-review", tokenFor(t, f.faculty), map[string]string{"decision": "approved"})
	if facultyRec.Code != http.StatusOK {
		t.Fatalf("faculty review status = %d, want %d; body = %s", facultyRec.Code, http.StatusOK, facultyRec.Body.String())
	}

	meRec := doJSON(t, r, http.MethodGet, "/consolidated-reports/me", studentToken, nil)
	var me map[string]any
	if err := json.Unmarshal(meRec.Body.Bytes(), &me); err != nil {
		t.Fatalf("decoding /me response: %v", err)
	}
	if me["agencyStatus"] != "approved" || me["facultyStatus"] != "approved" {
		t.Errorf("report = %+v, want both statuses approved", me)
	}
}

func TestResubmitFlow_RejectThenResubmitThenApprove(t *testing.T) {
	r, f := newTestRouter(t)
	studentToken := tokenFor(t, f.student)

	submitRec := doJSON(t, r, http.MethodPost, "/consolidated-reports", studentToken, map[string]string{
		"summary": "First draft.",
	})
	var created map[string]any
	if err := json.Unmarshal(submitRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding submit response: %v", err)
	}
	reportID := created["id"].(string)

	rejectRec := doJSON(t, r, http.MethodPost, "/consolidated-reports/"+reportID+"/agency-review", tokenFor(t, f.agencySup), map[string]string{"decision": "rejected"})
	if rejectRec.Code != http.StatusOK {
		t.Fatalf("reject status = %d, want %d; body = %s", rejectRec.Code, http.StatusOK, rejectRec.Body.String())
	}

	resubmitRec := doJSON(t, r, http.MethodPost, "/consolidated-reports/"+reportID+"/resubmit", studentToken, map[string]string{
		"summary": "Revised draft.",
	})
	if resubmitRec.Code != http.StatusOK {
		t.Fatalf("resubmit status = %d, want %d; body = %s", resubmitRec.Code, http.StatusOK, resubmitRec.Body.String())
	}
	var resubmitted map[string]any
	if err := json.Unmarshal(resubmitRec.Body.Bytes(), &resubmitted); err != nil {
		t.Fatalf("decoding resubmit response: %v", err)
	}
	if resubmitted["agencyStatus"] != "pending" {
		t.Fatalf("agencyStatus = %v, want pending after resubmit", resubmitted["agencyStatus"])
	}

	approveRec := doJSON(t, r, http.MethodPost, "/consolidated-reports/"+reportID+"/agency-review", tokenFor(t, f.agencySup), map[string]string{"decision": "approved"})
	if approveRec.Code != http.StatusOK {
		t.Fatalf("post-resubmit approve status = %d, want %d; body = %s", approveRec.Code, http.StatusOK, approveRec.Body.String())
	}
}

func TestResubmitHandler_RejectsBeforeAnyReview(t *testing.T) {
	r, f := newTestRouter(t)
	studentToken := tokenFor(t, f.student)

	submitRec := doJSON(t, r, http.MethodPost, "/consolidated-reports", studentToken, map[string]string{"summary": "First draft."})
	var created map[string]any
	if err := json.Unmarshal(submitRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding submit response: %v", err)
	}

	rec := doJSON(t, r, http.MethodPost, "/consolidated-reports/"+created["id"].(string)+"/resubmit", studentToken, map[string]string{"summary": "Too early."})
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}
