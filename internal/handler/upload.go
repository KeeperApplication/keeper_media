package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"

	"keeper.media/internal/service"
	"keeper.media/internal/util"
)

type UploadHandler struct {
	gcsService *service.GcsService
	logger     *slog.Logger
}

func NewUploadHandler(gcs *service.GcsService, logger *slog.Logger) *UploadHandler {
	return &UploadHandler{
		gcsService: gcs,
		logger:     logger,
	}
}

type contextKey string

const userContextKey = contextKey("username")

type PresignedURLRequest struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
}

type PresignedURLResponse struct {
	URL        string `json:"presignedUrl"`
	ObjectName string `json:"objectName"`
}

func (h *UploadHandler) GeneratePresignedURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		util.WriteJSONError(w, http.StatusMethodNotAllowed, "Method not allowed", h.logger)
		return
	}

	var req PresignedURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.WriteJSONError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	username, ok := r.Context().Value(userContextKey).(string)
	if !ok {
		h.logger.Error("Failed to retrieve username from request context after authentication")
		util.WriteJSONError(w, http.StatusInternalServerError, "Internal server error reading user identity", h.logger)
		return
	}

	if req.FileName == "" || req.ContentType == "" {
		util.WriteJSONError(w, http.StatusBadRequest, "fileName and contentType are required", h.logger)
		return
	}

	cleanFileName := filepath.Base(req.FileName)
	objectName := "avatars/" + username + "/" + cleanFileName

	url, err := h.gcsService.GenerateV4UploadURL(objectName, req.ContentType)
	if err != nil {
		h.logger.Error("Error generating signed URL", "error", err, "objectName", objectName)
		util.WriteJSONError(w, http.StatusInternalServerError, "Internal server error generating URL", h.logger)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(PresignedURLResponse{URL: url, ObjectName: objectName})
}
