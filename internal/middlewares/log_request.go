package middlewares

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

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
