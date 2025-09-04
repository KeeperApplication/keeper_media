package handler

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"keeper.media/internal/service"
	"keeper.media/internal/util"
)

type MediaHandler struct {
	gcsService *service.GcsService
	logger     *slog.Logger
}

func NewMediaHandler(gcs *service.GcsService, logger *slog.Logger) *MediaHandler {
	return &MediaHandler{
		gcsService: gcs,
		logger:     logger,
	}
}

func (h *MediaHandler) ServeMedia(w http.ResponseWriter, r *http.Request) {
	objectName := strings.TrimPrefix(r.URL.Path, "/media/")

	ctx := r.Context()
	reader, err := h.gcsService.ReadObject(ctx, objectName)
	if err != nil {
		h.logger.Warn("Could not get object from GCS", "object", objectName, "error", err)
		util.WriteJSONError(w, http.StatusNotFound, "File not found", h.logger)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", reader.Attrs.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")

	if _, err := io.Copy(w, reader); err != nil {
		h.logger.Error("Could not write response for object", "object", objectName, "error", err)
	}
}
