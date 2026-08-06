package middlewares

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"log"
	"net/http"
	"sync"

	"go.uber.org/zap/buffer"
)

var bufferPool = sync.Pool{
	New: func() any {
		return new(buffer.Buffer)
	},
}

// DecryptRequest decrypts encrypted request body.
func DecryptRequest(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			buf := bufferPool.Get().(*buffer.Buffer)
			buf.Reset()
			_, err := io.Copy(buf, r.Body)
			if err != nil {
				log.Fatal(nil)
			}
			decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, buf.Bytes())
			if err != nil {
				log.Fatal(err)
			}
			r.Body = io.NopCloser(bytes.NewReader(decryptedData))
			r.ContentLength = int64(len(decryptedData))
			next.ServeHTTP(w, r)
		})
	}
}
