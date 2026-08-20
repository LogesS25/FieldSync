package activities

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
	group := r.Group("/field-activities", auth.RequireAuth(jwtSecret), auth.RequireRole("student"))
	group.POST("", h.create)
	group.GET("", h.list)
}

type createFieldActivityRequest struct {
	ActivityDate string `json:"activityDate" binding:"required"`
	Description  string `json:"description" binding:"required"`
}

func (h *Handler) create(c *gin.Context) {
	var req createFieldActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	activityDate, err := db.ParseDate(req.ActivityDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "activityDate must be YYYY-MM-DD"})
		return
	}

	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	activity, err := h.service.Create(c.Request.Context(), studentID, activityDate, req.Description)
	if err != nil {
		if errors.Is(err, practicums.ErrNoActivePracticum) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create field activity"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(activity))
}

func (h *Handler) list(c *gin.Context) {
	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	activities, err := h.service.ListForStudent(c.Request.Context(), studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list field activities"})
		return
	}

	response := make([]gin.H, len(activities))
	for i, activity := range activities {
		response[i] = toResponse(activity)
	}
	c.JSON(http.StatusOK, response)
}

func toResponse(activity sqlcgen.FieldActivity) gin.H {
	return gin.H{
		"id":                 db.UUIDToString(activity.ID),
		"practicumId":        db.UUIDToString(activity.PracticumID),
		"activityDate":       db.DateToStringPtr(activity.ActivityDate),
		"description":        activity.Description,
		"verificationStatus": activity.VerificationStatus,
	}
}
