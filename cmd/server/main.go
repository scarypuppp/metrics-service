package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/scarypuppp/metrics-service/internal/config"
	"github.com/scarypuppp/metrics-service/internal/handler"
	"github.com/scarypuppp/metrics-service/internal/repository"
	"github.com/scarypuppp/metrics-service/internal/service"
	"go.uber.org/zap"
)

func main() {
	serverConfig, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	log.Printf("Listening on %s\n", serverConfig.Addr)

	storage := repository.NewMemStorage(serverConfig.FileStoragePath, serverConfig.StoreInterval == 0)

	if *serverConfig.Restore {
		if err := storage.RestoreFromFile(); err != nil {
			log.Println("failed to restore metrics:", err)
		}
	}
	if serverConfig.StoreInterval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(serverConfig.StoreInterval) * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if err := storage.SaveToFile(); err != nil {
					log.Println("failed to save metrics:", err)
				}
			}
		}()
	}

	metricService := service.MetricService{Storage: storage}
	router := handlers.GetAppRouter(metricService)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	srv := &http.Server{
		Addr:         serverConfig.Addr,
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

	log.Printf("Server started on %s", serverConfig.Addr)

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
