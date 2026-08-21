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
	// Used by the student's "form a practicum team" flow to pick a faculty
	// supervisor within their own university.
	r.GET("/faculty-supervisors/mine", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.listMyFacultySupervisors)
	// Used by the same flow to pick an agency supervisor once an agency has
	// been chosen.
	r.GET("/agency-supervisors", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.listAgencySupervisors)
}

func (h *Handler) me(c *gin.Context) {
	userID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	user, err := h.queries.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, toResponse(user))
}

func (h *Handler) listMyFacultySupervisors(c *gin.Context) {
	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	student, err := h.queries.GetUserByID(c.Request.Context(), studentID)
	if err != nil || !student.InstitutionID.Valid {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}

	faculty, err := h.queries.ListFacultySupervisorsForInstitution(c.Request.Context(), student.InstitutionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list faculty supervisors"})
		return
	}

	response := make([]gin.H, len(faculty))
	for i, u := range faculty {
		response[i] = toResponse(u)
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) listAgencySupervisors(c *gin.Context) {
	agencyID, err := db.ParseUUID(c.Query("agencyId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agencyId query param is required"})
		return
	}

	supervisors, err := h.queries.ListAgencySupervisorsForAgency(c.Request.Context(), agencyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list agency supervisors"})
		return
	}

	response := make([]gin.H, len(supervisors))
	for i, u := range supervisors {
		response[i] = toResponse(u)
	}
	c.JSON(http.StatusOK, response)
}

func toResponse(user sqlcgen.User) gin.H {
	return gin.H{
		"id":            db.UUIDToString(user.ID),
		"email":         user.Email,
		"fullName":      user.FullName,
		"role":          user.Role,
		"institutionId": db.UUIDToStringPtr(user.InstitutionID),
		"agencyId":      db.UUIDToStringPtr(user.AgencyID),
	}
}
