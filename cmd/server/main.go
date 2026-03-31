package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"

	"github.com/scarypuppp/metrics-service/internal/handler"
)

var serverOptions struct {
	addr string
}

func main() {
	serverOptions.addr = "localhost:8080"
	flag.Func("a", "server address host:port", func(flagValue string) error {
		expr, err := regexp.Compile(`^(localhost|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(\d{2,5})$`)
		if err != nil {
			return err
		}
		matches := expr.FindStringSubmatch(flagValue)
		if matches == nil {
			return fmt.Errorf("invalid address: %s", flagValue)
		}
		host := matches[1]
		port := matches[2]
		resultAddress := fmt.Sprintf("%s:%s", host, port)
		serverOptions.addr = resultAddress
		return nil
	})
	flag.Parse()
	fmt.Printf("Listening on %s\n", serverOptions.addr)
	router := handlers.GetAppRouter()
	log.Fatal(http.ListenAndServe(serverOptions.addr, router))
}
