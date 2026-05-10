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
	// Получение конфигурации
	serverConfig, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Инициализация логгера
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	// Инициализация sql.DB
	dbObj, err := db.NewDB(serverConfig.DatabaseDSN)
	if err != nil {
		logger.Error("create db connection failed", zap.Error(err))
	}
	defer dbObj.Close()

	// Инициализация репозитория метрик
	var storage repository.IMetricsStorage
	storageContext, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	if serverConfig.DatabaseDSN == "" {
		var opts []repository.Option
		if serverConfig.FileStoragePath != "" {
			opts = append(opts, repository.WithFile(
				storageContext,
				serverConfig.FileStoragePath,
				serverConfig.StoreInterval,
				*serverConfig.Restore,
			))
		}
		storage, err = repository.NewMemMetricsStorage(storageContext, opts...)
		if err != nil {
			logger.Fatal("Failed to setup memory storage", zap.Error(err))
		}
	} else {
		storage = repository.NewDBMetricsStorage(dbObj)
	}

	metricService := service.NewMetricService(storage)
	router := handlers.GetAppRouter(*metricService, dbObj)

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}
