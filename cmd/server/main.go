package main

import (
	"log"
	"net/http"

	"github.com/scarypuppp/metrics-service/internal/handler"
)

func main() {
	router := handlers.GetAppRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
}
