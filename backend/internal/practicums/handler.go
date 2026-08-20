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

func (h *Handler) RegisterRoutes(r gin.IRouter, jwtSecret string) {
	admin := r.Group("", auth.RequireAuth(jwtSecret), auth.RequireRole("administrator"))
	admin.POST("/practicums", h.createPracticum)
	admin.POST("/placements", h.createPlacement)
	admin.POST("/supervisor-assignments", h.createSupervisorAssignment)

	r.GET("/practicums/me", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.getMyPracticum)
	r.GET("/students", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor", "agency_supervisor"), h.listMyStudents)
}

type createPracticumRequest struct {
	StudentID     string `json:"studentId" binding:"required,uuid"`
	InstitutionID string `json:"institutionId" binding:"required,uuid"`
	StartDate     string `json:"startDate" binding:"required"`
	EndDate       string `json:"endDate"`
}

func (h *Handler) createPracticum(c *gin.Context) {
	var req createPracticumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentID, err := db.ParseUUID(req.StudentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid studentId"})
		return
	}
	institutionID, err := db.ParseUUID(req.InstitutionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid institutionId"})
		return
	}
	startDate, err := db.ParseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate must be YYYY-MM-DD"})
		return
	}
	endDate, err := db.ParseOptionalDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate must be YYYY-MM-DD"})
		return
	}

	practicum, err := h.service.CreatePracticum(c.Request.Context(), studentID, institutionID, startDate, endDate)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":            db.UUIDToString(practicum.ID),
		"studentId":     db.UUIDToString(practicum.StudentID),
		"institutionId": db.UUIDToString(practicum.InstitutionID),
		"status":        practicum.Status,
		"startDate":     db.DateToStringPtr(practicum.StartDate),
		"endDate":       db.DateToStringPtr(practicum.EndDate),
	})
}

type createPlacementRequest struct {
	PracticumID string `json:"practicumId" binding:"required,uuid"`
	AgencyID    string `json:"agencyId" binding:"required,uuid"`
	StartDate   string `json:"startDate" binding:"required"`
	EndDate     string `json:"endDate"`
}

func (h *Handler) createPlacement(c *gin.Context) {
	var req createPlacementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	practicumID, err := db.ParseUUID(req.PracticumID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid practicumId"})
		return
	}
	agencyID, err := db.ParseUUID(req.AgencyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agencyId"})
		return
	}
	startDate, err := db.ParseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate must be YYYY-MM-DD"})
		return
	}
	endDate, err := db.ParseOptionalDate(req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate must be YYYY-MM-DD"})
		return
	}

	placement, err := h.service.CreatePlacement(c.Request.Context(), practicumID, agencyID, startDate, endDate)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          db.UUIDToString(placement.ID),
		"practicumId": db.UUIDToString(placement.PracticumID),
		"agencyId":    db.UUIDToString(placement.AgencyID),
		"startDate":   db.DateToStringPtr(placement.StartDate),
		"endDate":     db.DateToStringPtr(placement.EndDate),
	})
}

type createSupervisorAssignmentRequest struct {
	PracticumID  string `json:"practicumId" binding:"required,uuid"`
	SupervisorID string `json:"supervisorId" binding:"required,uuid"`
}

func (h *Handler) createSupervisorAssignment(c *gin.Context) {
	var req createSupervisorAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	practicumID, err := db.ParseUUID(req.PracticumID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid practicumId"})
		return
	}
	supervisorID, err := db.ParseUUID(req.SupervisorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid supervisorId"})
		return
	}

	assignment, err := h.service.CreateSupervisorAssignment(c.Request.Context(), practicumID, supervisorID)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":           db.UUIDToString(assignment.ID),
		"practicumId":  db.UUIDToString(assignment.PracticumID),
		"supervisorId": db.UUIDToString(assignment.SupervisorID),
	})
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

func respondServiceError(c *gin.Context, err error) {
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
