package main

import (
	"fmt"
	"net/http"
)

func handleMetric(w http.ResponseWriter, req *http.Request) {
	metricType := req.PathValue("metricType")
	metricName := req.PathValue("metricName")
	metricValue := req.PathValue("metricValue")

	fmt.Println(metricType, metricName, metricValue)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/update/{metricType}/{metricName}/{metricValue}", handleMetric)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
