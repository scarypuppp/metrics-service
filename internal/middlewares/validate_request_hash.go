package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/scarypuppp/metrics-service/internal/utils/hash"
)

type responseCapture struct {
	http.ResponseWriter
	body       bytes.Buffer
	statusCode int
	headers    http.Header
}

func (rc *responseCapture) Header() http.Header {
	return rc.headers
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	return rc.body.Write(b)
}

func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
}

// ValidateRequestHash is a middleware that verifies the HashSHA256 header of requests
// and signs response bodies with the same HMAC key.
func ValidateRequestHash(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHash := r.Header.Get("HashSHA256")
			if receivedHash == "" {
				next.ServeHTTP(w, r)
				return
			}
			var buffer bytes.Buffer
			_, err := buffer.ReadFrom(r.Body)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			expectedHash := hash.GetHash(buffer.Bytes(), key)
			if receivedHash != expectedHash {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(&buffer)

			rc := &responseCapture{
				ResponseWriter: w,
				headers:        make(http.Header),
				statusCode:     http.StatusOK,
			}
			next.ServeHTTP(rc, r)

			// Вычисляем хэш тела ответа и проставляем заголовок
			responseHash := hash.GetHash(rc.body.Bytes(), key)
			rc.headers.Set("HashSHA256", responseHash)

			// Копируем заголовки в реальный ResponseWriter и отправляем ответ
			for k, v := range rc.headers {
				w.Header()[k] = v
			}
			w.WriteHeader(rc.statusCode)
			w.Write(rc.body.Bytes())
		})
	}
}
