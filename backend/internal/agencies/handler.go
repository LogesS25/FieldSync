// Package agencies manages the list of field placement organizations.
// Creation and listing are administrator-only — see the same rationale in
// internal/institutions.
package agencies

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
	queries *sqlcgen.Queries
}

func NewHandler(queries *sqlcgen.Queries) *Handler {
	return &Handler{queries: queries}
}

func (h *Handler) RegisterRoutes(r gin.IRouter, jwtSecret string) {
	group := r.Group("/agencies", auth.RequireAuth(jwtSecret), auth.RequireRole("administrator"))
	group.POST("", h.create)
	group.GET("", h.list)
}

type createAgencyRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) create(c *gin.Context) {
	var req createAgencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agency, err := h.queries.CreateAgency(c.Request.Context(), req.Name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "an agency with that name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create agency"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(agency))
}

func (h *Handler) list(c *gin.Context) {
	agencies, err := h.queries.ListAgencies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list agencies"})
		return
	}

	response := make([]gin.H, len(agencies))
	for i, ag := range agencies {
		response[i] = toResponse(ag)
	}
	c.JSON(http.StatusOK, response)
}

func toResponse(agency sqlcgen.Agency) gin.H {
	return gin.H{
		"id":   db.UUIDToString(agency.ID),
		"name": agency.Name,
	}
}
