package attendance

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

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
	r.POST("/attendance", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.create)
	r.GET("/attendance", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.list)
	r.GET("/attendance/pending", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor", "agency_supervisor"), h.listPending)
	r.POST("/attendance/:id/agency-review", auth.RequireAuth(jwtSecret), auth.RequireRole("agency_supervisor"), h.agencyReview)
	r.POST("/attendance/:id/faculty-review", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor"), h.facultyReview)
}

type createRequest struct {
	AttendanceDate string   `json:"attendanceDate" binding:"required"`
	Session        string   `json:"session" binding:"required,oneof=morning evening"`
	Hours          *float64 `json:"hours" binding:"omitempty,gt=0,lte=24"`
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	attendanceDate, err := db.ParseDate(req.AttendanceDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attendanceDate must be YYYY-MM-DD"})
		return
	}

	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	record, err := h.service.Create(c.Request.Context(), CreateInput{
		StudentID:      studentID,
		AttendanceDate: attendanceDate,
		Session:        sqlcgen.AttendanceSession(req.Session),
		Hours:          req.Hours,
	})
	if err != nil {
		if errors.Is(err, practicums.ErrNoActivePracticum) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "attendance already recorded for that date and session"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not record attendance"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(record))
}

func (h *Handler) list(c *gin.Context) {
	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	records, err := h.service.ListForStudent(c.Request.Context(), studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list attendance"})
		return
	}

	response := make([]gin.H, len(records))
	for i, record := range records {
		response[i] = toResponse(record)
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) listPending(c *gin.Context) {
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	records, err := h.service.ListPendingForSupervisor(c.Request.Context(), supervisorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list pending attendance"})
		return
	}

	response := make([]gin.H, len(records))
	for i, record := range records {
		response[i] = toResponse(record)
	}
	c.JSON(http.StatusOK, response)
}

type reviewRequest struct {
	Decision string `json:"decision" binding:"required,oneof=approved rejected"`
}

func (h *Handler) agencyReview(c *gin.Context) {
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recordID, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid record id"})
		return
	}
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	record, err := h.service.AgencyReview(c.Request.Context(), recordID, supervisorID, req.Decision == "approved")
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(record))
}

func (h *Handler) facultyReview(c *gin.Context) {
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recordID, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid record id"})
		return
	}
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	record, err := h.service.FacultyReview(c.Request.Context(), recordID, supervisorID, req.Decision == "approved")
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(record))
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrNotAssignedSupervisor):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrAgencyReviewFirst):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrAlreadyReviewedByRole):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "an unexpected error occurred"})
	}
}

func toResponse(record sqlcgen.AttendanceRecord) gin.H {
	var hours *float64
	if record.HoursLogged.Valid {
		h := db.NumericToFloat64(record.HoursLogged)
		hours = &h
	}
	return gin.H{
		"id":             db.UUIDToString(record.ID),
		"practicumId":    db.UUIDToString(record.PracticumID),
		"attendanceDate": db.DateToStringPtr(record.AttendanceDate),
		"session":        record.Session,
		"hours":          hours,
		"agencyStatus":   record.AgencyStatus,
		"facultyStatus":  record.FacultyStatus,
	}
}
