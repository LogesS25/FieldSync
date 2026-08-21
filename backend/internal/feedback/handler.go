package feedback

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r gin.IRouter, jwtSecret string) {
	r.POST("/feedback", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor", "agency_supervisor"), h.submit)
	r.GET("/feedback", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.listForMe)
	r.GET("/feedback/mine", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor", "agency_supervisor"), h.listMine)
}

type submitRequest struct {
	PracticumID   string `json:"practicumId" binding:"required,uuid"`
	WeekStartDate string `json:"weekStartDate" binding:"required"`
	Feedback      string `json:"feedback" binding:"required"`
}

func (h *Handler) submit(c *gin.Context) {
	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	practicumID, err := db.ParseUUID(req.PracticumID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid practicumId"})
		return
	}
	weekStart, err := db.ParseDate(req.WeekStartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "weekStartDate must be YYYY-MM-DD"})
		return
	}
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.service.Submit(c.Request.Context(), practicumID, supervisorID, weekStart, req.Feedback)
	if err != nil {
		if errors.Is(err, ErrNotAssignedSupervisor) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "you've already submitted feedback for this student for that week"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not submit feedback"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(entry))
}

func (h *Handler) listForMe(c *gin.Context) {
	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	entries, err := h.service.ListForStudent(c.Request.Context(), studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list feedback"})
		return
	}
	c.JSON(http.StatusOK, toResponseList(entries))
}

func (h *Handler) listMine(c *gin.Context) {
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	entries, err := h.service.ListFromSupervisor(c.Request.Context(), supervisorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list feedback"})
		return
	}
	c.JSON(http.StatusOK, toResponseList(entries))
}

func toResponse(entry sqlcgen.WeeklyFeedback) gin.H {
	return gin.H{
		"id":            db.UUIDToString(entry.ID),
		"practicumId":   db.UUIDToString(entry.PracticumID),
		"supervisorId":  db.UUIDToString(entry.SupervisorID),
		"weekStartDate": db.DateToStringPtr(entry.WeekStartDate),
		"feedback":      entry.Feedback,
	}
}

func toResponseList(entries []sqlcgen.WeeklyFeedback) []gin.H {
	response := make([]gin.H, len(entries))
	for i, e := range entries {
		response[i] = toResponse(e)
	}
	return response
}
