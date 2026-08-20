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
	group := r.Group("/attendance", auth.RequireAuth(jwtSecret), auth.RequireRole("student"))
	group.POST("", h.create)
	group.GET("", h.list)
	group.GET("/summary", h.summary)
}

type createAttendanceRequest struct {
	AttendanceDate string  `json:"attendanceDate" binding:"required"`
	Hours          float64 `json:"hours" binding:"required,gt=0,lte=24"`
}

func (h *Handler) create(c *gin.Context) {
	var req createAttendanceRequest
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

	record, err := h.service.Create(c.Request.Context(), studentID, attendanceDate, req.Hours)
	if err != nil {
		if errors.Is(err, practicums.ErrNoActivePracticum) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "attendance already recorded for that date"})
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

func (h *Handler) summary(c *gin.Context) {
	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	total, err := h.service.GetTotalHours(c.Request.Context(), studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not compute total hours"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"totalHours": total})
}

func toResponse(record sqlcgen.AttendanceRecord) gin.H {
	return gin.H{
		"id":                 db.UUIDToString(record.ID),
		"practicumId":        db.UUIDToString(record.PracticumID),
		"attendanceDate":     db.DateToStringPtr(record.AttendanceDate),
		"hours":              db.NumericToFloat64(record.HoursLogged),
		"verificationStatus": record.VerificationStatus,
	}
}
