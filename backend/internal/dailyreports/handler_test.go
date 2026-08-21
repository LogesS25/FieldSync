package dailyreports_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/dailyreports"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, testFixture) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	f := newFixture(t)
	r := gin.New()
	dailyreports.NewHandler(f.svc).RegisterRoutes(r, testSecret)
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

func doUpload(t *testing.T, r *gin.Engine, bearerToken, reportDate, filename, contentType string, fileBytes []byte) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if err := w.WriteField("reportDate", reportDate); err != nil {
		t.Fatalf("WriteField: %v", err)
	}

	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	partHeader.Set("Content-Type", contentType)
	part, err := w.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("CreatePart: %v", err)
	}
	if _, err := part.Write(fileBytes); err != nil {
		t.Fatalf("Write file part: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/daily-reports", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCreateHandler_RejectsNonPDF(t *testing.T) {
	r, f := newTestRouter(t)

	rec := doUpload(t, r, tokenFor(t, f.student), "2026-01-05", "report.txt", "text/plain", []byte("not a pdf"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCreateHandler_RequiresStudentRole(t *testing.T) {
	r, f := newTestRouter(t)

	rec := doUpload(t, r, tokenFor(t, f.faculty), "2026-01-05", "report.pdf", "application/pdf", []byte("%PDF-1.4"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestAgencyReviewHandler_RequiresAgencyRole(t *testing.T) {
	r, f := newTestRouter(t)

	createRec := doUpload(t, r, tokenFor(t, f.student), "2026-01-05", "report.pdf", "application/pdf", []byte("%PDF-1.4"))
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}

	rec := doJSON(t, r, http.MethodPost, "/daily-reports/"+created["id"].(string)+"/agency-review", tokenFor(t, f.faculty), map[string]string{
		"decision": "approved",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (only agency_supervisor role may agency-review); body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestFullApprovalFlow_AgencyThenFaculty(t *testing.T) {
	r, f := newTestRouter(t)
	studentToken := tokenFor(t, f.student)

	createRec := doUpload(t, r, studentToken, "2026-01-05", "report.pdf", "application/pdf", []byte("%PDF-1.4"))
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}
	reportID := created["id"].(string)

	facultyTooEarly := doJSON(t, r, http.MethodPost, "/daily-reports/"+reportID+"/faculty-review", tokenFor(t, f.faculty), map[string]string{"decision": "approved"})
	if facultyTooEarly.Code != http.StatusConflict {
		t.Fatalf("premature faculty review status = %d, want %d; body = %s", facultyTooEarly.Code, http.StatusConflict, facultyTooEarly.Body.String())
	}

	agencyRec := doJSON(t, r, http.MethodPost, "/daily-reports/"+reportID+"/agency-review", tokenFor(t, f.agencySup), map[string]string{"decision": "approved"})
	if agencyRec.Code != http.StatusOK {
		t.Fatalf("agency review status = %d, want %d; body = %s", agencyRec.Code, http.StatusOK, agencyRec.Body.String())
	}

	facultyRec := doJSON(t, r, http.MethodPost, "/daily-reports/"+reportID+"/faculty-review", tokenFor(t, f.faculty), map[string]string{"decision": "approved"})
	if facultyRec.Code != http.StatusOK {
		t.Fatalf("faculty review status = %d, want %d; body = %s", facultyRec.Code, http.StatusOK, facultyRec.Body.String())
	}

	listRec := doJSON(t, r, http.MethodGet, "/daily-reports", studentToken, nil)
	var reports []map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &reports); err != nil {
		t.Fatalf("decoding list response: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("len(reports) = %d, want 1", len(reports))
	}
	if reports[0]["agencyStatus"] != "approved" || reports[0]["facultyStatus"] != "approved" {
		t.Errorf("report = %+v, want both statuses approved", reports[0])
	}
}

func TestDownloadFileHandler_OwnerCanDownload(t *testing.T) {
	r, f := newTestRouter(t)
	studentToken := tokenFor(t, f.student)

	createRec := doUpload(t, r, studentToken, "2026-01-05", "report.pdf", "application/pdf", []byte("%PDF-1.4 hello"))
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}

	rec := doJSON(t, r, http.MethodGet, "/daily-reports/"+created["id"].(string)+"/file", studentToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestDownloadFileHandler_OtherStudentForbidden(t *testing.T) {
	r, f := newTestRouter(t)
	studentToken := tokenFor(t, f.student)

	createRec := doUpload(t, r, studentToken, "2026-01-05", "report.pdf", "application/pdf", []byte("%PDF-1.4"))
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}

	otherStudent := testutil.CreateTestUser(t, f.queries, sqlcgen.UserRoleStudent)
	rec := doJSON(t, r, http.MethodGet, "/daily-reports/"+created["id"].(string)+"/file", tokenFor(t, otherStudent), nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}
