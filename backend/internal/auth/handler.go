package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r gin.IRouter) {
	r.POST("/auth/register", h.register)
	r.POST("/auth/login", h.login)
	r.POST("/auth/refresh", h.refresh)
	r.POST("/auth/logout", h.logout)
}

type registerRequest struct {
	Email         string `json:"email" binding:"required,email"`
	Password      string `json:"password" binding:"required,min=8"`
	FullName      string `json:"fullName" binding:"required"`
	Role          string `json:"role" binding:"required"`
	InstitutionID string `json:"institutionId" binding:"omitempty,uuid"`
	AgencyID      string `json:"agencyId" binding:"omitempty,uuid"`
}

func (h *Handler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := sqlcgen.UserRole(req.Role)
	if !RegisterableRoles[role] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be one of: student, faculty_supervisor, agency_supervisor"})
		return
	}

	var institutionID, agencyID pgtype.UUID
	if req.InstitutionID != "" {
		var err error
		institutionID, err = db.ParseUUID(req.InstitutionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid institutionId"})
			return
		}
	}
	if req.AgencyID != "" {
		var err error
		agencyID, err = db.ParseUUID(req.AgencyID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid agencyId"})
			return
		}
	}

	session, err := h.service.Register(c.Request.Context(), RegisterInput{
		Email:         req.Email,
		Password:      req.Password,
		FullName:      req.FullName,
		Role:          role,
		InstitutionID: institutionID,
		AgencyID:      agencyID,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrEmailTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, ErrInstitutionRequired), errors.Is(err, ErrAgencyRequired),
			errors.Is(err, ErrInstitutionNotFound), errors.Is(err, ErrAgencyNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not register user"})
		}
		return
	}

	c.JSON(http.StatusCreated, sessionResponse(session))
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.service.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidCredentials.Error()})
		return
	}

	c.JSON(http.StatusOK, sessionResponse(session))
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (h *Handler) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidRefreshToken.Error()})
		return
	}

	c.JSON(http.StatusOK, sessionResponse(session))
}

func (h *Handler) logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not log out"})
		return
	}

	c.Status(http.StatusNoContent)
}

func sessionResponse(session Session) gin.H {
	return gin.H{
		"accessToken":  session.AccessToken,
		"refreshToken": session.RefreshToken,
		"user": gin.H{
			"id":            db.UUIDToString(session.User.ID),
			"email":         session.User.Email,
			"fullName":      session.User.FullName,
			"role":          session.User.Role,
			"institutionId": db.UUIDToStringPtr(session.User.InstitutionID),
			"agencyId":      db.UUIDToStringPtr(session.User.AgencyID),
		},
	}
}
