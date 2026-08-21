package practicums

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Practicum/Placement/SupervisorAssignment creation is no longer a direct
// admin action — it happens automatically when a practicum team request is
// mutually accepted (see internal/teamrequests). CreatePracticum,
// CreatePlacement, and CreateSupervisorAssignment on Service remain
// exported for that package to call.
func (h *Handler) RegisterRoutes(r gin.IRouter, jwtSecret string) {
	r.GET("/practicums/me", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.getMyPracticum)
	r.GET("/students", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor", "agency_supervisor"), h.listMyStudents)
}

func (h *Handler) getMyPracticum(c *gin.Context) {
	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	summary, err := h.service.GetSummaryForStudent(c.Request.Context(), studentID)
	if err != nil {
		if errors.Is(err, ErrNoActivePracticum) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load practicum"})
		return
	}

	var supervisors []map[string]any
	if err := json.Unmarshal(summary.Supervisors, &supervisors); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not parse supervisors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"practicumId":        db.UUIDToString(summary.PracticumID),
		"status":             summary.Status,
		"startDate":          db.DateToStringPtr(summary.StartDate),
		"endDate":            db.DateToStringPtr(summary.EndDate),
		"institutionName":    summary.InstitutionName,
		"agencyId":           db.UUIDToStringPtr(summary.AgencyID),
		"agencyName":         db.TextToStringPtr(summary.AgencyName),
		"placementStartDate": db.DateToStringPtr(summary.PlacementStartDate),
		"placementEndDate":   db.DateToStringPtr(summary.PlacementEndDate),
		"supervisors":        supervisors,
	})
}

func (h *Handler) listMyStudents(c *gin.Context) {
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	rows, err := h.service.ListForSupervisor(c.Request.Context(), supervisorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list students"})
		return
	}

	students := make([]gin.H, len(rows))
	for i, row := range rows {
		students[i] = gin.H{
			"studentId":       db.UUIDToString(row.StudentID),
			"studentName":     row.StudentName,
			"studentEmail":    row.StudentEmail,
			"practicumId":     db.UUIDToString(row.PracticumID),
			"practicumStatus": row.PracticumStatus,
			"startDate":       db.DateToStringPtr(row.PracticumStartDate),
			"endDate":         db.DateToStringPtr(row.PracticumEndDate),
			"institutionName": row.InstitutionName,
			"agencyId":        db.UUIDToStringPtr(row.AgencyID),
			"agencyName":      db.TextToStringPtr(row.AgencyName),
		}
	}

	c.JSON(http.StatusOK, students)
}

// RespondServiceError maps this package's sentinel errors to HTTP status
// codes. Exported so internal/teamrequests (which calls CreatePracticum /
// CreatePlacement / CreateSupervisorAssignment internally once a team
// request is mutually accepted) can reuse the same mapping instead of
// duplicating it.
func RespondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrStudentNotFound), errors.Is(err, ErrSupervisorNotFound), errors.Is(err, ErrPracticumNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrUserIsNotStudent), errors.Is(err, ErrUserIsNotSupervisor):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ErrAlreadyAssigned):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "an unexpected error occurred"})
	}
}
