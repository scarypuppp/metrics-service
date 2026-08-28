package middlewares

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// DecryptRequest decrypts encrypted request body.
func DecryptRequest(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body == nil || r.ContentLength == 0 {
				next.ServeHTTP(w, r)
				return
			}

			buf, ok := bufferPool.Get().(*bytes.Buffer)
			if !ok {
				buf = new(bytes.Buffer)
			}
			buf.Reset()
			defer bufferPool.Put(buf)

			if _, err := io.Copy(buf, r.Body); err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			if buf.Len() == 0 {
				next.ServeHTTP(w, r)
				return
			}

			decryptedData, err := rsa.DecryptPKCS1v15(nil, privateKey, buf.Bytes())
			if err != nil {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decryptedData))
			r.ContentLength = int64(len(decryptedData))

			next.ServeHTTP(w, r)
		})
	}
}
