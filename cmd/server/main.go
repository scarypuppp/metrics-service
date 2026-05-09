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
	"github.com/scarypuppp/metrics-service/internal/config/db"
	"github.com/scarypuppp/metrics-service/internal/handler"
	"github.com/scarypuppp/metrics-service/internal/repository"
	"github.com/scarypuppp/metrics-service/internal/service"
	"go.uber.org/zap"
)

func main() {
	serverConfig, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	dbConn, err := db.NewDB(serverConfig.DatabaseDSN)
	if err != nil {
		logger.Error("create db connection failed", zap.Error(err))
	}
	defer dbConn.Close()

	storage := repository.NewMemStorage(serverConfig.FileStoragePath, serverConfig.StoreInterval == 0)

	if *serverConfig.Restore {
		if err := storage.RestoreFromFile(); err != nil {
			logger.Error("failed to restore metrics", zap.Error(err))
		}
	}
	if serverConfig.StoreInterval > 0 {
		go func() {
			ticker := time.NewTicker(time.Duration(serverConfig.StoreInterval) * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				if err := storage.SaveToFile(); err != nil {
					logger.Error("failed to save metrics", zap.Error(err))
				}
			}
		}()
	}

	metricService := service.MetricService{Storage: storage}
	router := handlers.GetAppRouter(metricService, dbConn)

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
			logger.Error("run server failed", zap.Error(err))
		}
	}()

	logger.Info("Listening", zap.String("addr", serverConfig.Addr))
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")

}
