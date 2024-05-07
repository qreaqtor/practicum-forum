package middleware

import (
	"log/slog"
	"net/http"
)

// Если появляется ошибка, то пишет с помощью logger и возвращает код ответ 500
func Panic(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal server error", 500)
				logger.Error(
					"panic",
					"err", err,
					"url", r.URL.Path,
					"method", r.Method,
				)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
