package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"

	"github.com/scarypuppp/metrics-service/internal/audit"
	"github.com/scarypuppp/metrics-service/internal/repository"
	"github.com/scarypuppp/metrics-service/internal/service"
)

// newExampleRouter builds the application router backed by in-memory storage.
func newExampleRouter() http.Handler {
	storage, err := repository.NewMemMetricsStorage()
	if err != nil {
		panic(err)
	}
	metricService := service.NewMetricService(storage)
	return GetAppRouter("", netip.Prefix{}, nil, *metricService, audit.NewPublisher(), nil)
}

// doRequest sends a request to the router and returns the recorded response.
func doRequest(router http.Handler, method, target, body string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// ExampleUpdateMetricByURLHandler updates metrics via URL path parameters:
// POST /update/{metricType}/{metricName}/{metricValue}.
func ExampleUpdateMetricByURLHandler() {
	router := newExampleRouter()

	resp := doRequest(router, http.MethodPost, "/update/gauge/Alloc/123.45", "")
	fmt.Println(resp.Code, resp.Body.String())

	// Counter deltas accumulate between updates.
	doRequest(router, http.MethodPost, "/update/counter/PollCount/5", "")
	resp = doRequest(router, http.MethodPost, "/update/counter/PollCount/5", "")
	fmt.Println(resp.Code, resp.Body.String())

	// Output:
	// 200 123.45
	// 200 10
}

// ExampleUpdateMetricHandler updates a metric from a JSON body:
// POST /update/.
func ExampleUpdateMetricHandler() {
	router := newExampleRouter()

	resp := doRequest(router, http.MethodPost, "/update/",
		`{"id":"HeapAlloc","type":"gauge","value":42.5}`)
	fmt.Println(resp.Code)

	// Output:
	// 200
}

// ExampleUpdateMetricsHandler updates a batch of metrics from a JSON array:
// POST /updates/.
func ExampleUpdateMetricsHandler() {
	router := newExampleRouter()

	body := `[
		{"id":"Alloc","type":"gauge","value":42.5},
		{"id":"PollCount","type":"counter","delta":3}
	]`
	resp := doRequest(router, http.MethodPost, "/updates/", body)
	fmt.Println(resp.Code)

	resp = doRequest(router, http.MethodGet, "/value/counter/PollCount", "")
	fmt.Println(resp.Body.String())

	// Output:
	// 200
	// 3
}

// ExampleGetMetricByURLHandler reads a metric value via URL path parameters:
// GET /value/{metricType}/{metricName}.
func ExampleGetMetricByURLHandler() {
	router := newExampleRouter()
	doRequest(router, http.MethodPost, "/update/gauge/HeapAlloc/1024", "")

	resp := doRequest(router, http.MethodGet, "/value/gauge/HeapAlloc", "")
	fmt.Println(resp.Code, resp.Body.String())

	// Unknown metrics respond with 404.
	resp = doRequest(router, http.MethodGet, "/value/gauge/Unknown", "")
	fmt.Println(resp.Code)

	// Output:
	// 200 1024
	// 404
}

// ExampleGetMetricHandler reads a metric as JSON for a JSON request body:
// POST /value/.
func ExampleGetMetricHandler() {
	router := newExampleRouter()
	doRequest(router, http.MethodPost, "/update/",
		`{"id":"Alloc","type":"gauge","value":42.5}`)

	resp := doRequest(router, http.MethodPost, "/value/", `{"id":"Alloc","type":"gauge"}`)
	fmt.Println(resp.Code, resp.Body.String())

	// Output:
	// 200 {"id":"Alloc","type":"gauge","value":42.5}
}

// ExampleRetrieveMetricsHandler renders all stored metrics as an HTML page:
// GET /.
func ExampleRetrieveMetricsHandler() {
	router := newExampleRouter()
	doRequest(router, http.MethodPost, "/update/gauge/Alloc/42.5", "")
	doRequest(router, http.MethodPost, "/update/counter/PollCount/1", "")

	resp := doRequest(router, http.MethodGet, "/", "")
	fmt.Println(resp.Code, resp.Header().Get("Content-Type"))
	fmt.Println(strings.Contains(resp.Body.String(), "Alloc 42.5"))
	fmt.Println(strings.Contains(resp.Body.String(), "PollCount 1"))

	// Output:
	// 200 text/html; charset=utf-8
	// true
	// true
}
