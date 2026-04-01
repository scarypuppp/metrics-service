package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

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
	log.Printf("Listening on %s\n", serverOptions.addr)
	router := handlers.GetAppRouter()

	srv := &http.Server{
		Addr:         serverOptions.addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	log.Printf("Server started on %s", serverOptions.addr)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server stopped gracefully")

}
