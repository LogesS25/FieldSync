// Package httpserver wires the Gin router and top-level middleware.
package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fieldsync/backend/internal/agencies"
	"github.com/fieldsync/backend/internal/attendance"
	"github.com/fieldsync/backend/internal/auth"
	"github.com/fieldsync/backend/internal/config"
	"github.com/fieldsync/backend/internal/dailyreports"
	"github.com/fieldsync/backend/internal/db/sqlcgen"
	"github.com/fieldsync/backend/internal/feedback"
	"github.com/fieldsync/backend/internal/fieldworkcomponents"
	"github.com/fieldsync/backend/internal/institutions"
	"github.com/fieldsync/backend/internal/manuals"
	"github.com/fieldsync/backend/internal/notifications"
	"github.com/fieldsync/backend/internal/practicums"
	"github.com/fieldsync/backend/internal/reports"
	"github.com/fieldsync/backend/internal/storage"
	"github.com/fieldsync/backend/internal/teamrequests"
	"github.com/fieldsync/backend/internal/users"
)

func NewRouter(pool *pgxpool.Pool, cfg *config.Config) (*gin.Engine, error) {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Permissive CORS for local development so the Expo web target (a
	// different origin/port) can call the API. Tighten this to an explicit
	// origin allowlist before deploying anywhere real.
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
	}))

	r.GET("/health", healthHandler(pool))

	queries := sqlcgen.New(pool)
	authService := auth.NewService(queries, cfg.JWTSecret, cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	auth.NewHandler(authService).RegisterRoutes(r)
	users.NewHandler(queries).RegisterRoutes(r, cfg.JWTSecret)
	institutions.NewHandler(queries).RegisterRoutes(r, cfg.JWTSecret)
	agencies.NewHandler(queries).RegisterRoutes(r, cfg.JWTSecret)
	fieldworkcomponents.NewHandler(queries).RegisterRoutes(r, cfg.JWTSecret)

	notificationsService := notifications.NewService(queries)
	notifications.NewHandler(notificationsService).RegisterRoutes(r, cfg.JWTSecret)
	feedback.NewHandler(feedback.NewService(queries, notificationsService)).RegisterRoutes(r, cfg.JWTSecret)

	practicumsService := practicums.NewService(queries)
	practicums.NewHandler(practicumsService).RegisterRoutes(r, cfg.JWTSecret)
	attendance.NewHandler(attendance.NewService(queries, practicumsService, notificationsService)).RegisterRoutes(r, cfg.JWTSecret)
	reports.NewHandler(reports.NewService(queries, practicumsService, notificationsService)).RegisterRoutes(r, cfg.JWTSecret)
	teamrequests.NewHandler(teamrequests.NewService(queries, practicumsService, notificationsService)).RegisterRoutes(r, cfg.JWTSecret)

	store, err := storage.New(cfg.StorageDir)
	if err != nil {
		return nil, err
	}
	dailyreports.NewHandler(dailyreports.NewService(queries, practicumsService, store, notificationsService)).RegisterRoutes(r, cfg.JWTSecret)
	manuals.NewHandler(manuals.NewService(queries, store)).RegisterRoutes(r, cfg.JWTSecret)

	return r, nil
}

func healthHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "unreachable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	}
}
