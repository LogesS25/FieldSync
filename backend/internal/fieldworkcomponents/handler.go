// Package fieldworkcomponents manages the university-defined list students
// select from when forming a practicum team (business requirements §7).
// University-scoped and university-controlled like agencies (§4): create,
// update, and delete are administrator-only for now (no self-service
// University account exists yet — see internal/institutions doc comment for
// the same rationale), but the university's own control over add/remove/
// modify "at any time" is what update+delete implement.
package fieldworkcomponents

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
	admin := r.Group("/fieldwork-components", auth.RequireAuth(jwtSecret), auth.RequireRole("administrator"))
	admin.POST("", h.create)
	admin.GET("", h.list)
	admin.PATCH("/:id", h.update)
	admin.DELETE("/:id", h.delete)

	// Used by the student's "form a practicum team" flow to pick a
	// fieldwork component within their own university.
	r.GET("/fieldwork-components/mine", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.listMine)
}

type createRequest struct {
	Name          string `json:"name" binding:"required"`
	InstitutionID string `json:"institutionId" binding:"required,uuid"`
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	institutionID, err := db.ParseUUID(req.InstitutionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid institutionId"})
		return
	}

	component, err := h.queries.CreateFieldworkComponent(c.Request.Context(), sqlcgen.CreateFieldworkComponentParams{
		InstitutionID: institutionID,
		Name:          req.Name,
	})
	if err != nil {
		respondPgError(c, err, "could not create fieldwork component")
		return
	}

	c.JSON(http.StatusCreated, toResponse(component))
}

func (h *Handler) list(c *gin.Context) {
	components, err := h.queries.ListFieldworkComponents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list fieldwork components"})
		return
	}
	c.JSON(http.StatusOK, toResponseList(components))
}

func (h *Handler) listMine(c *gin.Context) {
	userID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	user, err := h.queries.GetUserByID(c.Request.Context(), userID)
	if err != nil || !user.InstitutionID.Valid {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	components, err := h.queries.ListFieldworkComponentsForInstitution(c.Request.Context(), user.InstitutionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list fieldwork components"})
		return
	}
	c.JSON(http.StatusOK, toResponseList(components))
}

type updateRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) update(c *gin.Context) {
	id, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	component, err := h.queries.UpdateFieldworkComponent(c.Request.Context(), sqlcgen.UpdateFieldworkComponentParams{
		ID:   id,
		Name: req.Name,
	})
	if err != nil {
		respondPgError(c, err, "could not update fieldwork component")
		return
	}
	c.JSON(http.StatusOK, toResponse(component))
}

func (h *Handler) delete(c *gin.Context) {
	id, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.queries.DeleteFieldworkComponent(c.Request.Context(), id); err != nil {
		respondPgError(c, err, "could not delete fieldwork component")
		return
	}
	c.Status(http.StatusNoContent)
}

func respondPgError(c *gin.Context, err error, fallback string) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			c.JSON(http.StatusConflict, gin.H{"error": "a fieldwork component with that name already exists for this university"})
			return
		case "23503":
			c.JSON(http.StatusConflict, gin.H{"error": "this fieldwork component is referenced by existing team requests and cannot be deleted"})
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
}

func toResponse(component sqlcgen.FieldworkComponent) gin.H {
	return gin.H{
		"id":            db.UUIDToString(component.ID),
		"name":          component.Name,
		"institutionId": db.UUIDToString(component.InstitutionID),
	}
}

func toResponseList(components []sqlcgen.FieldworkComponent) []gin.H {
	response := make([]gin.H, len(components))
	for i, c := range components {
		response[i] = toResponse(c)
	}
	return response
}
