package manuals

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
)

// maxUploadBytes caps the multipart body size Gin will parse — generous for
// a PDF guidance document, well below a DoS-sized payload.
const maxUploadBytes = 20 << 20 // 20 MiB

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r gin.IRouter, jwtSecret string) {
	admin := r.Group("/manuals", auth.RequireAuth(jwtSecret), auth.RequireRole("administrator"))
	admin.POST("", h.upload)
	admin.GET("", h.list)
	admin.DELETE("/:institutionId", h.delete)

	r.GET("/manuals/mine", auth.RequireAuth(jwtSecret), auth.RequireRole("student", "faculty_supervisor", "agency_supervisor"), h.getMine)
	// Authorization (caller belongs to this manual's university, or is an
	// administrator) is enforced in the service, not the role list — every
	// authenticated role may end up needing to view a manual.
	r.GET("/manuals/:id/file", auth.RequireAuth(jwtSecret), h.downloadFile)
}

func (h *Handler) upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes)

	institutionID, err := db.ParseUUID(c.PostForm("institutionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid institutionId"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	if fileHeader.Header.Get("Content-Type") != "application/pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file must be a PDF"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read uploaded file"})
		return
	}
	defer file.Close()

	uploaderID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	manual, err := h.service.Upload(c.Request.Context(), institutionID, uploaderID, fileHeader.Filename, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not upload manual"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(manual))
}

func (h *Handler) list(c *gin.Context) {
	manuals, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list manuals"})
		return
	}

	response := make([]gin.H, len(manuals))
	for i, manual := range manuals {
		response[i] = toResponse(manual)
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) delete(c *gin.Context) {
	institutionID, err := db.ParseUUID(c.Param("institutionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid institutionId"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), institutionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete manual"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) getMine(c *gin.Context) {
	callerID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	manual, err := h.service.GetForUser(c.Request.Context(), callerID)
	if err != nil {
		if errors.Is(err, ErrManualNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch manual"})
		return
	}
	c.JSON(http.StatusOK, toResponse(manual))
}

func (h *Handler) downloadFile(c *gin.Context) {
	manualID, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid manual id"})
		return
	}
	callerID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	role, _ := auth.CurrentUserRole(c)
	absPath, filename, err := h.service.GetFileForDownload(c.Request.Context(), manualID, callerID, role == "administrator")
	if err != nil {
		switch {
		case errors.Is(err, ErrManualNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrNotYourInstitution):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "an unexpected error occurred"})
		}
		return
	}

	c.FileAttachment(absPath, filename)
}

func toResponse(manual sqlcgen.Manual) gin.H {
	return gin.H{
		"id":            db.UUIDToString(manual.ID),
		"institutionId": db.UUIDToString(manual.InstitutionID),
		"filename":      manual.OriginalFilename,
		"updatedAt":     manual.UpdatedAt.Time,
	}
}
