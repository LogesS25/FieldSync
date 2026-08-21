package teamrequests_test

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
	"github.com/fieldsync/backend/internal/teamrequests"
)

const testSecret = "test-secret-do-not-use-in-prod"

func newTestRouter(t *testing.T) (*gin.Engine, testFixture) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	f := newFixture(t)
	r := gin.New()
	teamrequests.NewHandler(f.svc).RegisterRoutes(r, testSecret)
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

func baseCreateBody(f testFixture) map[string]string {
	return map[string]string{
		"agencyId":             db.UUIDToString(f.agency.ID),
		"facultySupervisorId":  db.UUIDToString(f.faculty.ID),
		"agencySupervisorId":   db.UUIDToString(f.agencySup.ID),
		"fieldworkComponentId": db.UUIDToString(f.component.ID),
		"fieldworkDescription": "Casework at a community clinic.",
		"startDate":            "2026-01-01",
	}
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

func TestCreateHandler_RequiresStudentRole(t *testing.T) {
	r, f := newTestRouter(t)

	rec := doJSON(t, r, http.MethodPost, "/team-requests", tokenFor(t, f.faculty), baseCreateBody(f))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestFullTeamFormationFlow(t *testing.T) {
	r, f := newTestRouter(t)
	studentToken := tokenFor(t, f.student)

	createRec := doJSON(t, r, http.MethodPost, "/team-requests", studentToken, baseCreateBody(f))
	if createRec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d; body = %s", createRec.Code, http.StatusCreated, createRec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}
	requestID := created["id"].(string)

	facultyPending := doJSON(t, r, http.MethodGet, "/team-requests/pending", tokenFor(t, f.faculty), nil)
	var facultyList []map[string]any
	if err := json.Unmarshal(facultyPending.Body.Bytes(), &facultyList); err != nil {
		t.Fatalf("decoding pending response: %v", err)
	}
	if len(facultyList) != 1 {
		t.Fatalf("len(facultyList) = %d, want 1", len(facultyList))
	}

	facultyRespond := doJSON(t, r, http.MethodPost, "/team-requests/"+requestID+"/respond", tokenFor(t, f.faculty), map[string]string{"decision": "accepted"})
	if facultyRespond.Code != http.StatusOK {
		t.Fatalf("faculty respond status = %d, want %d; body = %s", facultyRespond.Code, http.StatusOK, facultyRespond.Body.String())
	}

	agencyRespond := doJSON(t, r, http.MethodPost, "/team-requests/"+requestID+"/respond", tokenFor(t, f.agencySup), map[string]string{"decision": "accepted"})
	if agencyRespond.Code != http.StatusOK {
		t.Fatalf("agency respond status = %d, want %d; body = %s", agencyRespond.Code, http.StatusOK, agencyRespond.Body.String())
	}
	var agencyBody map[string]any
	if err := json.Unmarshal(agencyRespond.Body.Bytes(), &agencyBody); err != nil {
		t.Fatalf("decoding agency respond response: %v", err)
	}
	if agencyBody["formedPracticumId"] == nil {
		t.Fatal("expected formedPracticumId to be set once both supervisors accepted")
	}

	meRec := doJSON(t, r, http.MethodGet, "/team-requests/me", studentToken, nil)
	var mine []map[string]any
	if err := json.Unmarshal(meRec.Body.Bytes(), &mine); err != nil {
		t.Fatalf("decoding /team-requests/me response: %v", err)
	}
	if len(mine) != 1 || mine[0]["agencyDecision"] != "accepted" || mine[0]["facultyDecision"] != "accepted" {
		t.Errorf("mine = %+v, want one fully-accepted request", mine)
	}
}

func TestRespondHandler_RejectsInvalidDecision(t *testing.T) {
	r, f := newTestRouter(t)

	createRec := doJSON(t, r, http.MethodPost, "/team-requests", tokenFor(t, f.student), baseCreateBody(f))
	var created map[string]any
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decoding create response: %v", err)
	}

	rec := doJSON(t, r, http.MethodPost, "/team-requests/"+created["id"].(string)+"/respond", tokenFor(t, f.faculty), map[string]string{
		"decision": "maybe",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
