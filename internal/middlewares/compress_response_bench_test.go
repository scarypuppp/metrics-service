package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// полКБ повторяющегося текста — типичный сжимаемый ответ
var benchPayload = bytes.Repeat([]byte("metric gauge value 42.5\n"), 20)

func benchmarkCompress(b *testing.B, acceptGzip bool) {
	handler := CompressResponse(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(benchPayload)
	}))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if acceptGzip {
			req.Header.Set("Accept-Encoding", "gzip")
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkCompressResponseGzip(b *testing.B) {
	benchmarkCompress(b, true)
}

func BenchmarkCompressResponseNoGzip(b *testing.B) {
	benchmarkCompress(b, false)
}

func BenchmarkDecompressRequest(b *testing.B) {
	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	gz.Write(benchPayload)
	gz.Close()
	body := compressed.Bytes()

	handler := DecompressRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(io.Discard, r.Body)
	}))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Encoding", "gzip")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
