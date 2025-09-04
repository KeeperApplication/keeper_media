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

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/api/uploads/presigned-url", uploadHandler.GeneratePresignedURL)
	mux.HandleFunc("/media/", mediaHandler.ServeMedia)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{cfg.FrontEndURL},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	handler := c.Handler(mux)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
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
