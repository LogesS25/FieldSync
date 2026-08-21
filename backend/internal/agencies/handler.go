// Package agencies manages the list of field placement organizations.
// Agencies are university-scoped (business requirements §5 — a university
// owns its own agency list). Creation/full listing are administrator-only —
// see internal/institutions for the same rationale (no self-service
// University account exists yet). `GET /agencies/mine` lets a student or
// faculty supervisor browse the agencies available to their own
// institution, e.g. when forming a practicum team request.
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
	admin := r.Group("/agencies", auth.RequireAuth(jwtSecret), auth.RequireRole("administrator"))
	admin.POST("", h.create)
	admin.GET("", h.list)
	admin.PATCH("/:id", h.update)
	admin.DELETE("/:id", h.delete)

	r.GET("/agencies/mine", auth.RequireAuth(jwtSecret), h.listMine)

	// Unauthenticated: the registration screen needs to show an agency
	// picker before the user has an account (agency supervisors register
	// against a specific agency).
	r.GET("/public/agencies", h.list)
}

type createAgencyRequest struct {
	Name          string `json:"name" binding:"required"`
	InstitutionID string `json:"institutionId" binding:"required,uuid"`
}

func (h *Handler) create(c *gin.Context) {
	var req createAgencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	institutionID, err := db.ParseUUID(req.InstitutionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid institutionId"})
		return
	}

	agency, err := h.queries.CreateAgency(c.Request.Context(), sqlcgen.CreateAgencyParams{
		Name:          req.Name,
		InstitutionID: institutionID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				c.JSON(http.StatusConflict, gin.H{"error": "an agency with that name already exists"})
				return
			case "23503":
				c.JSON(http.StatusBadRequest, gin.H{"error": "institution not found"})
				return
			}
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

func (h *Handler) listMine(c *gin.Context) {
	userID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	user, err := h.queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user in token"})
		return
	}
	if !user.InstitutionID.Valid {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	agencies, err := h.queries.ListAgenciesForInstitution(c.Request.Context(), user.InstitutionID)
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

type updateAgencyRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) update(c *gin.Context) {
	id, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateAgencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agency, err := h.queries.UpdateAgency(c.Request.Context(), sqlcgen.UpdateAgencyParams{ID: id, Name: req.Name})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "an agency with that name already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update agency"})
		return
	}
	c.JSON(http.StatusOK, toResponse(agency))
}

func (h *Handler) delete(c *gin.Context) {
	id, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.queries.DeleteAgency(c.Request.Context(), id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			c.JSON(http.StatusConflict, gin.H{"error": "this agency is referenced by existing practicum records and cannot be deleted"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete agency"})
		return
	}
	c.Status(http.StatusNoContent)
}

func toResponse(agency sqlcgen.Agency) gin.H {
	return gin.H{
		"id":            db.UUIDToString(agency.ID),
		"name":          agency.Name,
		"institutionId": db.UUIDToString(agency.InstitutionID),
	}
}
