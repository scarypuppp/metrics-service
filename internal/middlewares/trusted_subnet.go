package middlewares

import (
	"net/http"
	"net/netip"
)

// TrustedSubnet returns a middleware that rejects requests whose IP is outside the given prefix.
func TrustedSubnet(prefix netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestIP := GetCtxRequestIP(r.Context())
			ip, err := netip.ParseAddr(requestIP)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			if !prefix.Contains(ip) {
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
