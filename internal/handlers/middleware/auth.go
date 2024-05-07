package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"forum/internal/models"
	"strings"
)

// Check - проверка токена
type authManager interface {
	Check(string) (*models.Author, error)
}

type authMiddleware struct {
	auth   authManager
	logger *slog.Logger
}

func NewAuthMiddleware(auth authManager, logger *slog.Logger) *authMiddleware {
	return &authMiddleware{
		auth:   auth,
		logger: logger,
	}
}

// Возвращает хендлер, проверяющий токен
func (am *authMiddleware) GetHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenIn := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		author, err := am.auth.Check(tokenIn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			am.logger.Error(
				"auth middleware",
				"token", tokenIn,
				"err", err.Error(),
				"url", r.URL.Path,
				"method", r.Method,
			)
			return
		}
		ctx := context.WithValue(r.Context(), models.CtxKey("user"), author)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
