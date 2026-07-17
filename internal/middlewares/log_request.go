package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// LogRequest is a middleware that logs the URI, method and duration of each request.
func LogRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		handler.ServeHTTP(w, r)

		duration := time.Since(start)

		zap.S().Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
		)
	})
}
