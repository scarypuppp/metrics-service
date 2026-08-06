package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/scarypuppp/metrics-service/internal/agent"
	"go.uber.org/zap"
)

func main() {

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	agentConfig, err := agent.GetConfig()
	if err != nil {
		logger.Fatal("read agent config failed", zap.Error(err))
	}

	client := resty.NewWithClient(&http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	newAgent := agent.NewAgent(
		agent.NewCollector(),
		agent.NewSender(client, agentConfig.ServerAddr, agentConfig.Key),
		*logger,
		agentConfig.PoolInterval,
		agentConfig.ReportInterval,
	)

	log.Println("Agent started")
	newAgent.RunCtx(ctx, agentConfig.RateLimit)
	log.Println("Agent stopped")
}
