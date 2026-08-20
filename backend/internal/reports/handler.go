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
	group := r.Group("/weekly-reports", auth.RequireAuth(jwtSecret), auth.RequireRole("student"))
	group.POST("", h.submit)
	group.GET("", h.list)
}

type submitWeeklyReportRequest struct {
	WeekStartDate string `json:"weekStartDate" binding:"required"`
	WeekEndDate   string `json:"weekEndDate" binding:"required"`
	Summary       string `json:"summary" binding:"required"`
}

func (h *Handler) submit(c *gin.Context) {
	var req submitWeeklyReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	weekStart, err := db.ParseDate(req.WeekStartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "weekStartDate must be YYYY-MM-DD"})
		return
	}
	weekEnd, err := db.ParseDate(req.WeekEndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "weekEndDate must be YYYY-MM-DD"})
		return
	}

	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.Submit(c.Request.Context(), studentID, weekStart, weekEnd, req.Summary)
	if err != nil {
		if errors.Is(err, practicums.ErrNoActivePracticum) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "a report for that week has already been submitted"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not submit weekly report"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(report))
}

func (h *Handler) list(c *gin.Context) {
	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	reportList, err := h.service.ListForStudent(c.Request.Context(), studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list weekly reports"})
		return
	}

	response := make([]gin.H, len(reportList))
	for i, report := range reportList {
		response[i] = toResponse(report)
	}
	c.JSON(http.StatusOK, response)
}

func toResponse(report sqlcgen.WeeklyReport) gin.H {
	return gin.H{
		"id":            db.UUIDToString(report.ID),
		"practicumId":   db.UUIDToString(report.PracticumID),
		"weekStartDate": db.DateToStringPtr(report.WeekStartDate),
		"weekEndDate":   db.DateToStringPtr(report.WeekEndDate),
		"summary":       report.Summary,
		"status":        report.Status,
	}
}
