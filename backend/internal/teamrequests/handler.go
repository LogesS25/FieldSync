package teamrequests

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/practicums"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r gin.IRouter, jwtSecret string) {
	r.POST("/team-requests", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.create)
	r.GET("/team-requests/me", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.listMine)
	r.GET("/team-requests/pending", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor", "agency_supervisor"), h.listPending)
	r.POST("/team-requests/:id/respond", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor", "agency_supervisor"), h.respond)
}

type createRequest struct {
	AgencyID             string `json:"agencyId" binding:"required,uuid"`
	FacultySupervisorID  string `json:"facultySupervisorId" binding:"required,uuid"`
	AgencySupervisorID   string `json:"agencySupervisorId" binding:"required,uuid"`
	FieldworkDescription string `json:"fieldworkDescription" binding:"required"`
	StartDate            string `json:"startDate" binding:"required"`
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	agencyID, err := db.ParseUUID(req.AgencyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agencyId"})
		return
	}
	facultyID, err := db.ParseUUID(req.FacultySupervisorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid facultySupervisorId"})
		return
	}
	agencySupID, err := db.ParseUUID(req.AgencySupervisorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agencySupervisorId"})
		return
	}
	startDate, err := db.ParseDate(req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate must be YYYY-MM-DD"})
		return
	}

	request, err := h.service.CreateRequest(c.Request.Context(), CreateRequestInput{
		StudentID:            studentID,
		AgencyID:             agencyID,
		FacultySupervisorID:  facultyID,
		AgencySupervisorID:   agencySupID,
		FieldworkDescription: req.FieldworkDescription,
		StartDate:            startDate,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(request))
}

func (h *Handler) listMine(c *gin.Context) {
	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	requests, err := h.service.ListForStudent(c.Request.Context(), studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list team requests"})
		return
	}

	response := make([]gin.H, len(requests))
	for i, req := range requests {
		response[i] = toResponse(req)
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) listPending(c *gin.Context) {
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	requests, err := h.service.ListForSupervisor(c.Request.Context(), supervisorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list team requests"})
		return
	}

	response := make([]gin.H, len(requests))
	for i, req := range requests {
		response[i] = toResponse(req)
	}
	c.JSON(http.StatusOK, response)
}

type respondRequest struct {
	Decision string `json:"decision" binding:"required,oneof=accepted rejected"`
}

func (h *Handler) respond(c *gin.Context) {
	var req respondRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestID, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}

	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	roleVal, _ := c.Get(auth.ContextRoleKey)
	accept := req.Decision == "accepted"

	var updated sqlcgen.PracticumTeamRequest
	if roleVal == string(sqlcgen.UserRoleFacultySupervisor) {
		updated, err = h.service.RespondAsFaculty(c.Request.Context(), requestID, supervisorID, accept)
	} else {
		updated, err = h.service.RespondAsAgency(c.Request.Context(), requestID, supervisorID, accept)
	}
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(updated))
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrStudentNotFound), errors.Is(err, ErrAgencyNotFound),
		errors.Is(err, ErrFacultyNotFound), errors.Is(err, ErrAgencySupNotFound),
		errors.Is(err, ErrRequestNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrUserIsNotStudent), errors.Is(err, ErrStudentHasNoUniversity),
		errors.Is(err, ErrAgencyWrongUniversity), errors.Is(err, ErrFacultyWrongRole),
		errors.Is(err, ErrFacultyWrongUniversity), errors.Is(err, ErrAgencySupWrongRole),
		errors.Is(err, ErrAgencySupWrongAgency):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ErrNotYourRequest):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrAlreadyDecided):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, practicums.ErrAlreadyAssigned):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "an unexpected error occurred"})
	}
}

func toResponse(r sqlcgen.PracticumTeamRequest) gin.H {
	return gin.H{
		"id":                   db.UUIDToString(r.ID),
		"studentId":            db.UUIDToString(r.StudentID),
		"agencyId":             db.UUIDToString(r.AgencyID),
		"facultySupervisorId":  db.UUIDToString(r.FacultySupervisorID),
		"agencySupervisorId":   db.UUIDToString(r.AgencySupervisorID),
		"fieldworkDescription": r.FieldworkDescription,
		"startDate":            db.DateToStringPtr(r.StartDate),
		"facultyDecision":      r.FacultyDecision,
		"agencyDecision":       r.AgencyDecision,
		"formedPracticumId":    db.UUIDToStringPtr(r.FormedPracticumID),
	}
}
