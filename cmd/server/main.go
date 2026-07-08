package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/scarypuppp/metrics-service/internal/audit"
	"github.com/scarypuppp/metrics-service/internal/config"
	handlers "github.com/scarypuppp/metrics-service/internal/handler"
	"github.com/scarypuppp/metrics-service/internal/infrastructure/postgres"
	"github.com/scarypuppp/metrics-service/internal/repository"
	"github.com/scarypuppp/metrics-service/internal/service"
	"go.uber.org/zap"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func runMigrations(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("create migrate: %w", err)
	}
	defer m.Close()

	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migrations: %w", err)
	}
	return nil
}

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
	dbObj, err := postgres.NewDB(serverConfig.DatabaseDSN)
	if err != nil {
		logger.Fatal("db object creation failed", zap.Error(err))
	}
	defer dbObj.Close()

	// Инициализация репозитория метрик
	var storage repository.MetricsStorage
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
		storage, err = repository.NewMemMetricsStorage(opts...)
		if err != nil {
			logger.Fatal("failed to setup memory storage", zap.Error(err))
		}
		logger.Info("Using memory storage", zap.Bool("with_file", len(opts) == 1))
	} else {
		if err := runMigrations(serverConfig.DatabaseDSN); err != nil {
			log.Fatal(err)
		}
		storage = repository.NewDBMetricsStorage(dbObj)
		logger.Info("Using database storage")
	}
	metricService := service.NewMetricService(storage)

	publisher := audit.NewPublisher()
	subsCtx, subsCancel := context.WithCancel(context.Background())
	defer subsCancel()

	var subscribers []audit.Subscriber // interface { Wait(); Stop() }

	if serverConfig.AuditFile != "" {
		fileSub, err := audit.NewFileSubscriber(subsCtx, publisher, "file", serverConfig.AuditFile)
		if err != nil {
			logger.Error("error starting file audit subscriber", zap.Error(err))
		} else {
			subscribers = append(subscribers, fileSub)
		}
	}
	if serverConfig.AuditURL != "" {
		urlSub, err := audit.NewURLSubscriber(subsCtx, publisher, "url", serverConfig.AuditURL)
		if err != nil {
			logger.Error("error starting url audit subscriber", zap.Error(err))
		} else {
			subscribers = append(subscribers, urlSub)
		}
	}

	router := handlers.GetAppRouter(
		serverConfig.Key,
		*metricService,
		publisher,
		dbObj,
	)

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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	publisher.Close()

	// 3. Ждём подписчиков с таймаутом
	subsDone := make(chan struct{})
	go func() {
		for _, s := range subscribers {
			s.Wait()
		}
		close(subsDone)
	}()

	select {
	case <-subsDone:
		logger.Info("audit subscribers finished")
	case <-time.After(10 * time.Second):
		logger.Warn("audit subscribers timed out, forcing stop")
		for _, s := range subscribers {
			s.Stop()
		}
	}

	logger.Info("Server stopped gracefully")
}
