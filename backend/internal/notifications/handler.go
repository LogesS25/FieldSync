package notifications

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

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
	group := r.Group("/notifications", auth.RequireAuth(jwtSecret))
	group.GET("", h.list)
	group.POST("/:id/read", h.markRead)
	group.POST("/read-all", h.markAllRead)
}

func (h *Handler) list(c *gin.Context) {
	userID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	items, err := h.service.ListForUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list notifications"})
		return
	}

	response := make([]gin.H, len(items))
	for i, item := range items {
		response[i] = toResponse(item)
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) markRead(c *gin.Context) {
	id, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}
	userID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	notification, err := h.service.MarkRead(c.Request.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotificationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrNotYourNotification):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "an unexpected error occurred"})
		}
		return
	}
	c.JSON(http.StatusOK, toResponse(notification))
}

func (h *Handler) markAllRead(c *gin.Context) {
	userID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.MarkAllRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not mark notifications read"})
		return
	}
	c.Status(http.StatusNoContent)
}

func toResponse(n sqlcgen.Notification) gin.H {
	var readAt *time.Time
	if n.ReadAt.Valid {
		readAt = &n.ReadAt.Time
	}
	return gin.H{
		"id":        db.UUIDToString(n.ID),
		"message":   n.Message,
		"readAt":    readAt,
		"createdAt": n.CreatedAt.Time,
	}
}
