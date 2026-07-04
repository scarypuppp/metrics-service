package middlewares

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type contextKey string

const RequestIPKey contextKey = "requestIP"

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
