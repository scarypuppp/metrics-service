package middlewares

import (
	"net/http"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (response *loggingResponseWriter) Write(bytes []byte) (int, error) {
	size, err := response.ResponseWriter.Write(bytes)
	response.responseData.size += size
	return size, err
}

func (response *loggingResponseWriter) WriteHeader(statusCode int) {
	response.ResponseWriter.WriteHeader(statusCode)
	response.responseData.status = statusCode
}

func LogResponse(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		handler.ServeHTTP(&lw, r)

		zap.S().Infow("response",
			"status", responseData.status,
			"size", responseData.size,
		)
	})
}
