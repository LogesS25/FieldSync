package dailyreports

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/db"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/practicums"
)

// maxUploadBytes caps the multipart body size Gin will parse — well above
// a scanned handwritten page as a PDF, well below a DoS-sized payload.
const maxUploadBytes = 20 << 20 // 20 MiB

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r gin.IRouter, jwtSecret string) {
	r.POST("/daily-reports", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.create)
	r.GET("/daily-reports", auth.RequireAuth(jwtSecret), auth.RequireRole("student"), h.list)
	r.GET("/daily-reports/pending", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor", "agency_supervisor"), h.listPending)
	r.GET("/daily-reports/:id/file", auth.RequireAuth(jwtSecret), auth.RequireRole("student", "faculty_supervisor", "agency_supervisor"), h.downloadFile)
	r.POST("/daily-reports/:id/agency-review", auth.RequireAuth(jwtSecret), auth.RequireRole("agency_supervisor"), h.agencyReview)
	r.POST("/daily-reports/:id/faculty-review", auth.RequireAuth(jwtSecret), auth.RequireRole("faculty_supervisor"), h.facultyReview)
}

func (h *Handler) create(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes)

	reportDate, err := db.ParseDate(c.PostForm("reportDate"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reportDate must be YYYY-MM-DD"})
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

	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.Create(c.Request.Context(), studentID, reportDate, fileHeader.Filename, file)
	if err != nil {
		if errors.Is(err, practicums.ErrNoActivePracticum) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "a daily report already exists for that date"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not submit daily report"})
		return
	}

	c.JSON(http.StatusCreated, toResponse(report))
}

func (h *Handler) list(c *gin.Context) {
	studentID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	reports, err := h.service.ListForStudent(c.Request.Context(), studentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list daily reports"})
		return
	}

	response := make([]gin.H, len(reports))
	for i, report := range reports {
		response[i] = toResponse(report)
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) listPending(c *gin.Context) {
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	reports, err := h.service.ListPendingForSupervisor(c.Request.Context(), supervisorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list pending daily reports"})
		return
	}

	response := make([]gin.H, len(reports))
	for i, report := range reports {
		response[i] = toResponse(report)
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) downloadFile(c *gin.Context) {
	reportID, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}
	callerID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	role, _ := auth.CurrentUserRole(c)
	absPath, filename, err := h.service.GetFileForDownload(c.Request.Context(), reportID, callerID, role == "student")
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.FileAttachment(absPath, filename)
}

type reviewRequest struct {
	Decision string `json:"decision" binding:"required,oneof=approved rejected"`
}

func (h *Handler) agencyReview(c *gin.Context) {
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reportID, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.AgencyReview(c.Request.Context(), reportID, supervisorID, req.Decision == "approved")
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(report))
}

func (h *Handler) facultyReview(c *gin.Context) {
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reportID, err := db.ParseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}
	supervisorID, err := auth.CurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	report, err := h.service.FacultyReview(c.Request.Context(), reportID, supervisorID, req.Decision == "approved")
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, toResponse(report))
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrReportNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrNotAssignedSupervisor), errors.Is(err, ErrNotYourReport):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrAgencyReviewFirst), errors.Is(err, ErrAlreadyReviewedByRole):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "an unexpected error occurred"})
	}
}

func toResponse(report sqlcgen.DailyReport) gin.H {
	return gin.H{
		"id":            db.UUIDToString(report.ID),
		"practicumId":   db.UUIDToString(report.PracticumID),
		"reportDate":    db.DateToStringPtr(report.ReportDate),
		"filename":      report.OriginalFilename,
		"agencyStatus":  report.AgencyStatus,
		"facultyStatus": report.FacultyStatus,
	}
}
