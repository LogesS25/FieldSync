// Package institutions manages the list of educational institutions
// students and faculty supervisors belong to. Creation and listing are
// administrator-only: institutions are reference data maintained by staff,
// not something individual users create. There is no Admin UI yet (Phase 9
// builds it) — until then, these routes are exercised directly (curl/DBeaver
// to create an administrator user) — see docs/ARCHITECTURE.md §8.
package institutions

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
	group := r.Group("/institutions", auth.RequireAuth(jwtSecret), auth.RequireRole("administrator"))
	group.POST("", h.create)
	group.GET("", h.list)

	// Unauthenticated: the registration screen needs to show a university
	// picker before the user has an account.
	r.GET("/public/institutions", h.list)
}

type createInstitutionRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) create(c *gin.Context) {
	var req createInstitutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	institution, err := h.queries.CreateInstitution(c.Request.Context(), req.Name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "an institution with that name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create institution"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(institution))
}

func (h *Handler) list(c *gin.Context) {
	institutions, err := h.queries.ListInstitutions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list institutions"})
		return
	}

	response := make([]gin.H, len(institutions))
	for i, inst := range institutions {
		response[i] = toResponse(inst)
	}
	c.JSON(http.StatusOK, response)
}

func toResponse(institution sqlcgen.Institution) gin.H {
	return gin.H{
		"id":   db.UUIDToString(institution.ID),
		"name": institution.Name,
	}
}
