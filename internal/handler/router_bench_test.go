package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/scarypuppp/metrics-service/internal/audit"
	"github.com/scarypuppp/metrics-service/internal/repository"
	"github.com/scarypuppp/metrics-service/internal/service"
)

func newBenchRouter(b *testing.B) http.Handler {
	b.Helper()
	storage, err := repository.NewMemMetricsStorage()
	if err != nil {
		b.Fatal(err)
	}
	metricService := service.NewMetricService(storage)
	return GetAppRouter("", netip.Prefix{}, nil, *metricService, audit.NewPublisher(), nil)
}

func benchRequest(b *testing.B, router http.Handler, method, target string, body []byte) {
	b.Helper()
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		var req *http.Request
		b.StopTimer()
		if body != nil {
			reader.Reset(body)
			req = httptest.NewRequest(method, target, reader)
			req.Header.Set("Content-Type", "application/json")
		} else {
			req = httptest.NewRequest(method, target, nil)
		}
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()
		b.StartTimer()
		router.ServeHTTP(rec, req)
		if rec.Code >= http.StatusBadRequest {
			b.Fatalf("unexpected status %d: %s", rec.Code, rec.Body.String())
		}
	}
}

func BenchmarkRouterUpdateMetricJSON(b *testing.B) {
	b.StopTimer()
	router := newBenchRouter(b)
	body := []byte(`{"id":"benchGauge","type":"gauge","value":42.5}`)
	b.StartTimer()
	benchRequest(b, router, http.MethodPost, "/update/", body)
}

func BenchmarkRouterUpdateMetricURL(b *testing.B) {
	b.StopTimer()
	router := newBenchRouter(b)
	b.StartTimer()
	benchRequest(b, router, http.MethodPost, "/update/gauge/benchGauge/42.5", nil)
}

func BenchmarkRouterUpdateMetricsBatch(b *testing.B) {
	b.StopTimer()
	router := newBenchRouter(b)
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i := 0; i < 50; i++ {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(&buf, `{"id":"gauge%d","type":"gauge","value":%d.5}`, i, i)
	}
	buf.WriteByte(']')
	b.StartTimer()
	benchRequest(b, router, http.MethodPost, "/updates/", buf.Bytes())
}

func BenchmarkRouterGetMetricJSON(b *testing.B) {
	b.StopTimer()
	router := newBenchRouter(b)
	seed := httptest.NewRequest(http.MethodPost, "/update/gauge/benchGauge/42.5", nil)
	router.ServeHTTP(httptest.NewRecorder(), seed)

	body := []byte(`{"id":"benchGauge","type":"gauge"}`)
	b.StartTimer()
	benchRequest(b, router, http.MethodPost, "/value/", body)
}

func BenchmarkRouterGetMetricByURL(b *testing.B) {
	b.StopTimer()
	router := newBenchRouter(b)
	seed := httptest.NewRequest(http.MethodPost, "/update/gauge/benchGauge/42.5", nil)
	router.ServeHTTP(httptest.NewRecorder(), seed)
	b.StartTimer()
	benchRequest(b, router, http.MethodGet, "/value/gauge/benchGauge", nil)
}

func BenchmarkRouterRetrieveAllMetrics(b *testing.B) {
	b.StopTimer()
	router := newBenchRouter(b)
	for i := 0; i < 100; i++ {
		seed := httptest.NewRequest(http.MethodPost,
			fmt.Sprintf("/update/gauge/gauge%d/%d.5", i, i), nil)
		router.ServeHTTP(httptest.NewRecorder(), seed)
	}
	b.StartTimer()
	benchRequest(b, router, http.MethodGet, "/", nil)
}
