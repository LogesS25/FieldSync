package practicums_test

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
	"github.com/fieldsync/backend/internal/practicums"
	"github.com/fieldsync/backend/internal/testutil"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, *sqlcgen.Queries) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	queries := testutil.NewTestQueries(t)
	r := gin.New()
	practicums.NewHandler(practicums.NewService(queries)).RegisterRoutes(r, testSecret)
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

func TestCreatePracticumHandler_RequiresAdminRole(t *testing.T) {
	r, queries := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	rec := doJSON(t, r, http.MethodPost, "/practicums", tokenFor(t, student), map[string]string{
		"studentId":     db.UUIDToString(student.ID),
		"institutionId": db.UUIDToString(student.ID), // irrelevant, should be rejected before use
		"startDate":     "2026-01-01",
	})

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (non-admin must be rejected); body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCreatePracticumHandler_RejectsInvalidUUID(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)

	rec := doJSON(t, r, http.MethodPost, "/practicums", tokenFor(t, admin), map[string]string{
		"studentId":     "not-a-uuid",
		"institutionId": "also-not-a-uuid",
		"startDate":     "2026-01-01",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestCreatePracticumHandler_RejectsInvalidDate(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	institution := testutil.CreateTestInstitution(t, queries)

	rec := doJSON(t, r, http.MethodPost, "/practicums", tokenFor(t, admin), map[string]string{
		"studentId":     db.UUIDToString(student.ID),
		"institutionId": db.UUIDToString(institution.ID),
		"startDate":     "01/01/2026",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestGetMyPracticumHandler_RequiresStudentRole(t *testing.T) {
	r, queries := newTestRouter(t)
	supervisor := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleFacultySupervisor)

	rec := doJSON(t, r, http.MethodGet, "/practicums/me", tokenFor(t, supervisor), nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestGetMyPracticumHandler_NoActivePracticumReturns404(t *testing.T) {
	r, queries := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	rec := doJSON(t, r, http.MethodGet, "/practicums/me", tokenFor(t, student), nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestListStudentsHandler_RequiresSupervisorRole(t *testing.T) {
	r, queries := newTestRouter(t)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)

	rec := doJSON(t, r, http.MethodGet, "/students", tokenFor(t, student), nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestFullPracticumFlow_AdminSetsUpStudentSeesItSupervisorSeesIt(t *testing.T) {
	r, queries := newTestRouter(t)
	admin := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAdministrator)
	student := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleStudent)
	supervisor := testutil.CreateTestUser(t, queries, sqlcgen.UserRoleAgencySupervisor)
	institution := testutil.CreateTestInstitution(t, queries)
	agency := testutil.CreateTestAgency(t, queries)
	adminToken := tokenFor(t, admin)

	practicumRec := doJSON(t, r, http.MethodPost, "/practicums", adminToken, map[string]string{
		"studentId":     db.UUIDToString(student.ID),
		"institutionId": db.UUIDToString(institution.ID),
		"startDate":     "2026-01-01",
	})
	if practicumRec.Code != http.StatusCreated {
		t.Fatalf("create practicum status = %d, want %d; body = %s", practicumRec.Code, http.StatusCreated, practicumRec.Body.String())
	}
	var practicumBody map[string]any
	if err := json.Unmarshal(practicumRec.Body.Bytes(), &practicumBody); err != nil {
		t.Fatalf("decoding practicum response: %v", err)
	}
	practicumID := practicumBody["id"].(string)

	placementRec := doJSON(t, r, http.MethodPost, "/placements", adminToken, map[string]string{
		"practicumId": practicumID,
		"agencyId":    db.UUIDToString(agency.ID),
		"startDate":   "2026-01-01",
	})
	if placementRec.Code != http.StatusCreated {
		t.Fatalf("create placement status = %d, want %d; body = %s", placementRec.Code, http.StatusCreated, placementRec.Body.String())
	}

	assignmentRec := doJSON(t, r, http.MethodPost, "/supervisor-assignments", adminToken, map[string]string{
		"practicumId":  practicumID,
		"supervisorId": db.UUIDToString(supervisor.ID),
	})
	if assignmentRec.Code != http.StatusCreated {
		t.Fatalf("create assignment status = %d, want %d; body = %s", assignmentRec.Code, http.StatusCreated, assignmentRec.Body.String())
	}

	meRec := doJSON(t, r, http.MethodGet, "/practicums/me", tokenFor(t, student), nil)
	if meRec.Code != http.StatusOK {
		t.Fatalf("GET /practicums/me status = %d, want %d; body = %s", meRec.Code, http.StatusOK, meRec.Body.String())
	}
	var meBody map[string]any
	if err := json.Unmarshal(meRec.Body.Bytes(), &meBody); err != nil {
		t.Fatalf("decoding /practicums/me response: %v", err)
	}
	if meBody["institutionName"] != institution.Name {
		t.Errorf("institutionName = %v, want %v", meBody["institutionName"], institution.Name)
	}
	if meBody["agencyName"] != agency.Name {
		t.Errorf("agencyName = %v, want %v", meBody["agencyName"], agency.Name)
	}

	studentsRec := doJSON(t, r, http.MethodGet, "/students", tokenFor(t, supervisor), nil)
	if studentsRec.Code != http.StatusOK {
		t.Fatalf("GET /students status = %d, want %d; body = %s", studentsRec.Code, http.StatusOK, studentsRec.Body.String())
	}
	var students []map[string]any
	if err := json.Unmarshal(studentsRec.Body.Bytes(), &students); err != nil {
		t.Fatalf("decoding /students response: %v", err)
	}
	if len(students) != 1 {
		t.Fatalf("len(students) = %d, want 1", len(students))
	}
	if students[0]["studentId"] != db.UUIDToString(student.ID) {
		t.Errorf("studentId = %v, want %v", students[0]["studentId"], db.UUIDToString(student.ID))
	}
}
