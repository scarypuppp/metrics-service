package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"
)

func DecompressRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			handler.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer gz.Close()
		r.Body = gz
		handler.ServeHTTP(w, r)
	})
}
