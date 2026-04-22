package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"myproject/internal/checker"
	"myproject/internal/db"
	"myproject/internal/monitor"
)

func main() {
	// Structured JSON logger — every log line is machine-readable JSON
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Create a cancellable context that cancels on SIGINT / SIGTERM
	// This single context drives graceful shutdown for everything
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Read DATABASE_URL from env; fall back to local dev default
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:secret@localhost:5432/postgres"
	}

	// Try to connect to Postgres; fall back to in-memory repo if unavailable
	var repo monitor.Repository
	pool, err := db.New(ctx, connStr)
	if err != nil {
		slog.Warn("main: postgres not available, falling back to memory repo", "err", err)
		repo = monitor.NewMemoryRepo()
	} else {
		defer pool.Close()
		slog.Info("main: connected to postgres")
		repo = monitor.NewPostgresRepo(pool)
	}

	// Build the HTTP handler and register all REST routes
	h := monitor.NewHandler(repo)
	router := gin.Default()

	api := router.Group("/api/v1")
	{
		api.POST("/monitors", h.Create)              // Create a new monitor
		api.GET("/monitors", h.List)                 // List all monitors
		api.GET("/monitors/:id", h.GetByID)          // Get one monitor by ID
		api.PUT("/monitors/:id", h.Update)           // Update a monitor
		api.DELETE("/monitors/:id", h.Delete)        // Delete a monitor
		api.GET("/monitors/:id/checks", h.ListChecks) // Get check history
	}

	// Health check endpoint — used by load balancers / Docker healthcheck
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Prometheus metrics endpoint — scraped by Prometheus every 5s
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Start the background health-checker worker pool in its own goroutine
	chk := checker.New(repo)
	go chk.Start(ctx)

	// Build the HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start serving in background
	go func() {
		slog.Info("main: server listening", "addr", ":8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("main: server crashed", "err", err)
			cancel() // propagate failure to all goroutines
		}
	}()

	// Block until OS signal received (Ctrl+C or SIGTERM from Docker)
	<-ctx.Done()
	slog.Info("main: shutdown signal received")

	// Give in-flight HTTP requests up to 10 s to finish
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("main: graceful shutdown failed", "err", err)
	}

	slog.Info("main: stopped cleanly")
}
