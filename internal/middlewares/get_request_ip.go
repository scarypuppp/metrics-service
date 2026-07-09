package middlewares

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type contextKey string

// RequestIPKey is the context key under which GetRequestIP stores the client IP address.
const RequestIPKey contextKey = "requestIP"

// GetRequestIP is a middleware that resolves the client IP from headers or the remote address
// and stores it in the request context under RequestIPKey.
func GetRequestIP(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var result string

		for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} {
			if addresses := r.Header.Get(header); addresses != "" {
				result = strings.TrimSpace(strings.Split(addresses, ",")[0])
				break
			}
		}

		if result == "" {
			if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				result = ip
			} else {
				result = r.RemoteAddr
			}
		}

		ctx := context.WithValue(r.Context(), RequestIPKey, result)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
