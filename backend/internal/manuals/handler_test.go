package manuals_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/manuals"
	"github.com/fieldsync/backend/internal/storage"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, *sqlcgen.Queries) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	queries := testutil.NewTestQueries(t)
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	r := gin.New()
	manuals.NewHandler(manuals.NewService(queries, store)).RegisterRoutes(r, testSecret)
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

func doUpload(t *testing.T, r *gin.Engine, bearerToken, institutionID, filename, contentType string, fileBytes []byte) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	if err := w.WriteField("institutionId", institutionID); err != nil {
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

	req := httptest.NewRequest(http.MethodPost, "/manuals", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func createUserWithInstitution(t *testing.T, queries *sqlcgen.Queries, role sqlcgen.UserRole, institutionID pgtype.UUID) sqlcgen.User {
	t.Helper()
	user, err := queries.CreateUser(t.Context(), sqlcgen.CreateUserParams{
		Email:         fmt.Sprintf("%s-%d@example.com", role, time.Now().UnixNano()),
		PasswordHash:  "irrelevant",
		Role:          role,
		FullName:      "Test User",
		InstitutionID: institutionID,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	return user
}

func createAgencySupervisor(t *testing.T, queries *sqlcgen.Queries, agencyID pgtype.UUID) sqlcgen.User {
	t.Helper()
	user, err := queries.CreateUser(t.Context(), sqlcgen.CreateUserParams{
		Email:        fmt.Sprintf("agsup-%d@example.com", time.Now().UnixNano()),
		PasswordHash: "irrelevant",
		Role:         sqlcgen.UserRoleAgencySupervisor,
		FullName:     "Test Agency Supervisor",
		AgencyID:     agencyID,
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	return user
}

func TestUploadHandler_RequiresAdmin(t *testing.T) {
	r, queries := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)

	rec := doUpload(t, r, tokenFor(t, student), db.UUIDToString(institution.ID), "manual.pdf", "application/pdf", []byte("%PDF-1.4"))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestUploadHandler_RejectsNonPDF(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institution := testutil.CreateTestInstitution(t, queries)

	rec := doUpload(t, r, tokenFor(t, admin), db.UUIDToString(institution.ID), "manual.txt", "text/plain", []byte("not a pdf"))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestUploadHandler_ReplacesExistingManual(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institution := testutil.CreateTestInstitution(t, queries)
	adminToken := tokenFor(t, admin)

	first := doUpload(t, r, adminToken, db.UUIDToString(institution.ID), "v1.pdf", "application/pdf", []byte("%PDF-1.4 v1"))
	if first.Code != http.StatusCreated {
		t.Fatalf("first upload status = %d, want %d; body = %s", first.Code, http.StatusCreated, first.Body.String())
	}

	second := doUpload(t, r, adminToken, db.UUIDToString(institution.ID), "v2.pdf", "application/pdf", []byte("%PDF-1.4 v2"))
	if second.Code != http.StatusCreated {
		t.Fatalf("second upload status = %d, want %d; body = %s", second.Code, http.StatusCreated, second.Body.String())
	}

	listRec := doJSON(t, r, http.MethodGet, "/manuals", adminToken, nil)
	var list []map[string]any
	if err := json.Unmarshal(listRec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decoding list response: %v", err)
	}
	count := 0
	var filename any
	for _, item := range list {
		if item["institutionId"] == db.UUIDToString(institution.ID) {
			count++
			filename = item["filename"]
		}
	}
	if count != 1 {
		t.Fatalf("count of manuals for institution = %d, want 1 (replace, not accumulate)", count)
	}
	if filename != "v2.pdf" {
		t.Errorf("filename = %v, want v2.pdf (latest upload)", filename)
	}
}

func TestGetMine_StudentSeesOwnInstitutionManual(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institution := testutil.CreateTestInstitution(t, queries)
	doUpload(t, r, tokenFor(t, admin), db.UUIDToString(institution.ID), "manual.pdf", "application/pdf", []byte("%PDF-1.4"))

	student := createUserWithInstitution(t, queries, sqlcgen.UserRoleStudent, institution.ID)

	rec := doJSON(t, r, http.MethodGet, "/manuals/mine", tokenFor(t, student), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestGetMine_AgencySupervisorResolvesViaAgency(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institution := testutil.CreateTestInstitution(t, queries)
	agency := testutil.CreateTestAgency(t, queries, institution.ID)
	doUpload(t, r, tokenFor(t, admin), db.UUIDToString(institution.ID), "manual.pdf", "application/pdf", []byte("%PDF-1.4"))

	agSup := createAgencySupervisor(t, queries, agency.ID)

	rec := doJSON(t, r, http.MethodGet, "/manuals/mine", tokenFor(t, agSup), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestGetMine_NotFoundWhenNoManualUploaded(t *testing.T) {
	r, queries := newTestRouter(t)
	institution := testutil.CreateTestInstitution(t, queries)
	student := createUserWithInstitution(t, queries, sqlcgen.UserRoleStudent, institution.ID)

	rec := doJSON(t, r, http.MethodGet, "/manuals/mine", tokenFor(t, student), nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestDownloadFile_OtherUniversityForbidden(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institutionA := testutil.CreateTestInstitution(t, queries)
	institutionB := testutil.CreateTestInstitution(t, queries)
	uploadRec := doUpload(t, r, tokenFor(t, admin), db.UUIDToString(institutionA.ID), "manual.pdf", "application/pdf", []byte("%PDF-1.4"))
	var created map[string]any
	if err := json.Unmarshal(uploadRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding upload response: %v", err)
	}

	studentB := createUserWithInstitution(t, queries, sqlcgen.UserRoleStudent, institutionB.ID)

	rec := doJSON(t, r, http.MethodGet, "/manuals/"+created["id"].(string)+"/file", tokenFor(t, studentB), nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestDownloadFile_OwnUniversitySucceeds(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	institution := testutil.CreateTestInstitution(t, queries)
	uploadRec := doUpload(t, r, tokenFor(t, admin), db.UUIDToString(institution.ID), "manual.pdf", "application/pdf", []byte("%PDF-1.4 hello"))
	var created map[string]any
	if err := json.Unmarshal(uploadRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding upload response: %v", err)
	}

	student := createUserWithInstitution(t, queries, sqlcgen.UserRoleStudent, institution.ID)

	rec := doJSON(t, r, http.MethodGet, "/manuals/"+created["id"].(string)+"/file", tokenFor(t, student), nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestDeleteHandler_RequiresAdmin(t *testing.T) {
	r, queries := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)

	rec := doJSON(t, r, http.MethodDelete, "/manuals/"+db.UUIDToString(institution.ID), tokenFor(t, student), nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}
