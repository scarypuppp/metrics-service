package middlewares

import (
	"compress/gzip"
	"net/http"
	"strings"
	"sync"
)

// gzipReaderPool переиспользует gzip.Reader между запросами:
// буферы декомпрессора аллоцируются один раз, а не на каждое сжатое тело.
var gzipReaderPool = sync.Pool{
	New: func() any {
		return new(gzip.Reader)
	},
}

func DecompressRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			handler.ServeHTTP(w, r)
			return
		}

		gz := gzipReaderPool.Get().(*gzip.Reader)
		if err := gz.Reset(r.Body); err != nil {
			gzipReaderPool.Put(gz)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			gz.Close()
			gzipReaderPool.Put(gz)
		}()
		r.Body = gz
		handler.ServeHTTP(w, r)
	})
}
