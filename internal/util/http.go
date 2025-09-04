package util

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func WriteJSONError(w http.ResponseWriter, status int, message string, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		logger.Error("failed to write json error response", "error", err)
	}
}
