package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"keeper.media/internal/config"
	"keeper.media/internal/handler"
	"keeper.media/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()

	gcsService, err := service.NewGcsService(cfg)
	if err != nil {
		logger.Error("Failed to create GCS service", "error", err)
		os.Exit(1)
	}

	uploadHandler := handler.NewUploadHandler(gcsService, logger)
	mediaHandler := handler.NewMediaHandler(gcsService, logger)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{cfg.FrontEndURL},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	r.Use(c.Handler)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	})

	r.Post("/api/uploads/presigned-url", uploadHandler.GeneratePresignedURL)

	r.Mount("/media", http.StripPrefix("/media", http.HandlerFunc(mediaHandler.ServeMedia)))

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("Starting media service", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Could not start server", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdownCtx.Done()

	logger.Info("Shutdown signal received, starting graceful shutdown")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(timeoutCtx); err != nil {
		logger.Error("Graceful shutdown failed", "error", err)
	} else {
		logger.Info("Server shutdown gracefully")
	}
}
