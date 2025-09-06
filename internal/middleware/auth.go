package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"keeper.media/internal/auth"
	"keeper.media/internal/util"
)

type contextKey string

const UserContextKey = contextKey("username")

func AuthMiddleware(jwtSecret string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				util.WriteJSONError(w, http.StatusUnauthorized, "Authorization header required", logger)
				return
			}

			tokenParts := strings.Split(authHeader, "Bearer ")
			if len(tokenParts) != 2 || tokenParts[1] == "" {
				util.WriteJSONError(w, http.StatusUnauthorized, "Invalid token format. Expected Bearer token.", logger)
				return
			}
			tokenString := tokenParts[1]

			username, err := auth.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				logger.Warn("Failed to validate token", "error", err)
				util.WriteJSONError(w, http.StatusUnauthorized, "Invalid token", logger)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
