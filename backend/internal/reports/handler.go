package reports

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
	r.POST("/consolidated-reports", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.submit)
	r.GET("/consolidated-reports/me", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.getMine)
	r.GET("/consolidated-reports/pending", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor", "agency_supervisor"), h.listPending)
	r.POST("/consolidated-reports/:id/agency-review", auth.RequireAuth(jwtSecret), auth.RequireRole("agency_supervisor"), h.agencyReview)
	r.POST("/consolidated-reports/:id/faculty-review", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor"), h.facultyReview)
	r.POST("/consolidated-reports/:id/resubmit", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.resubmit)
}

type submitRequest struct {
	Summary string `json:"summary" binding:"required"`
}

func (h *Handler) submit(c *gin.Context) {
	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.Submit(c.Request.Context(), studentID, req.Summary)
	if err != nil {
		if errors.Is(err, practicums.ErrNoActivePracticum) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "a consolidated report has already been submitted for this practicum"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not submit consolidated report"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(report))
}

func (h *Handler) getMine(c *gin.Context) {
	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.GetForStudent(c.Request.Context(), studentID)
	if err != nil {
		if errors.Is(err, ErrReportNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load consolidated report"})
		return
	}

	c.JSON(http.StatusOK, toResponse(report))
}

func (h *Handler) resubmit(c *gin.Context) {
	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reportID, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}
	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.Resubmit(c.Request.Context(), reportID, studentID, req.Summary)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(report))
}

func (h *Handler) listPending(c *gin.Context) {
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	reportList, err := h.service.ListPendingForSupervisor(c.Request.Context(), supervisorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list pending reports"})
		return
	}

	response := make([]gin.H, len(reportList))
	for i, report := range reportList {
		response[i] = toResponse(report)
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
	reportID, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.AgencyReview(c.Request.Context(), reportID, supervisorID, req.Decision == "approved")
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(report))
}

func (h *Handler) facultyReview(c *gin.Context) {
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reportID, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.FacultyReview(c.Request.Context(), reportID, supervisorID, req.Decision == "approved")
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(report))
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrReportNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrNotAssignedSupervisor), errors.Is(err, ErrNotYourReport):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrAgencyReviewFirst), errors.Is(err, ErrAlreadyReviewedByRole), errors.Is(err, ErrNotRejected):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "an unexpected error occurred"})
	}
}

func toResponse(report sqlcgen.ConsolidatedReport) gin.H {
	return gin.H{
		"id":            db.UUIDToString(report.ID),
		"practicumId":   db.UUIDToString(report.PracticumID),
		"summary":       report.Summary,
		"agencyStatus":  report.AgencyStatus,
		"facultyStatus": report.FacultyStatus,
		"submittedAt":   report.SubmittedAt.Time,
	}
}
