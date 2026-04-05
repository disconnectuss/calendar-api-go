package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-go/internal/auth"
	"api-go/internal/calendar"
	"api-go/internal/config"
	"api-go/internal/middleware"
	"api-go/internal/tasks"
	"api-go/internal/webhooks"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Middleware stack
	r.Use(middleware.LoggingMiddleware())
	r.Use(securityHeaders())
	r.Use(middleware.RateLimitMiddleware())
	r.Use(middleware.CORSMiddleware(cfg.App.AllowedOrigins))

	// Token storage
	tokenStorage := auth.NewTokenStorage()

	// Auth middleware
	authMw := middleware.AuthMiddleware(cfg, tokenStorage)

	// API v1 group
	v1 := r.Group("/v1")

	// Root endpoint
	v1.GET("/", rootHandler)

	// Auth routes (no auth middleware)
	authHandler := auth.NewHandler(cfg, tokenStorage)
	authHandler.RegisterRoutes(v1)

	// Protected routes
	protected := v1.Group("")
	protected.Use(authMw)

	// Calendar routes
	calendarService := calendar.NewService()
	calendarHandler := calendar.NewHandler(cfg, calendarService)
	calendarHandler.RegisterRoutes(protected)

	// Tasks routes
	tasksService := tasks.NewService()
	tasksHandler := tasks.NewHandler(cfg, tasksService)
	tasksHandler.RegisterRoutes(protected)

	// Webhook routes (mixed auth)
	webhookService := webhooks.NewService()
	webhookHandler := webhooks.NewHandler(cfg, webhookService)
	webhookHandler.RegisterRoutes(v1, authMw)

	// Server
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		slog.Info("server starting", "port", cfg.App.Port, "env", cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}
	slog.Info("server exited")
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Next()
	}
}

func rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":    "Google Calendar & Tasks API",
		"version": "1.0.0",
		"endpoints": gin.H{
			"auth": gin.H{
				"GET /v1/auth/google":          "Start OAuth2 flow",
				"GET /v1/auth/google/callback":  "OAuth2 callback",
			},
			"calendar": gin.H{
				"GET /v1/calendar/events":         "List events (paginated)",
				"GET /v1/calendar/events/:id":     "Get event by ID",
				"POST /v1/calendar/events/general": "Create general event",
				"POST /v1/calendar/events/private": "Create private event",
				"PATCH /v1/calendar/events/:id":    "Update event",
				"DELETE /v1/calendar/events/:id":   "Delete event",
			},
			"tasks": gin.H{
				"GET /v1/tasks/lists":                                    "List task lists",
				"POST /v1/tasks/lists":                                   "Create task list",
				"GET /v1/tasks":                                          "List default tasks",
				"POST /v1/tasks":                                         "Create default task",
				"GET /v1/tasks/lists/:taskListId/tasks":                  "List tasks in list",
				"POST /v1/tasks/lists/:taskListId/tasks":                 "Create task in list",
				"GET /v1/tasks/lists/:taskListId/tasks/:taskId":          "Get task",
				"PATCH /v1/tasks/lists/:taskListId/tasks/:taskId":        "Update task",
				"DELETE /v1/tasks/lists/:taskListId/tasks/:taskId":       "Delete task",
				"POST /v1/tasks/lists/:taskListId/tasks/:taskId/complete":   "Complete task",
				"POST /v1/tasks/lists/:taskListId/tasks/:taskId/uncomplete": "Uncomplete task",
			},
			"webhooks": gin.H{
				"POST /v1/webhooks/subscribe":     "Subscribe to calendar notifications",
				"POST /v1/webhooks/unsubscribe":   "Unsubscribe from notifications",
				"POST /v1/webhooks/notifications": "Receive notifications (no auth)",
				"GET /v1/webhooks/channels":       "List active channels",
			},
		},
	})
}
