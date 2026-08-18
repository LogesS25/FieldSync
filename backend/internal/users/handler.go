package users

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

type Handler struct {
	queries *sqlcgen.Queries
}

func NewHandler(queries *sqlcgen.Queries) *Handler {
	return &Handler{queries: queries}
}

func (h *Handler) RegisterRoutes(r gin.IRouter, jwtSecret string) {
	r.GET("/users/me", auth.RequireAuth(jwtSecret), h.me)
}

func (h *Handler) me(c *gin.Context) {
	userIDStr, _ := c.Get(auth.ContextUserIDKey)
	userID, err := db.ParseUUID(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user in token"})
		return
	}

	user, err := h.queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       db.UUIDToString(user.ID),
		"email":    user.Email,
		"fullName": user.FullName,
		"role":     user.Role,
	})
}
