package main

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/scarypuppp/metrics-service/internal/agent"
)

func main() {

	agentConfig, err := agent.GetConfig()
	if err != nil {
		panic(err)
	}

	client := resty.NewWithClient(&http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	})
	newAgent := agent.NewAgent(
		agent.NewCollector(),
		agent.NewSender(client, agentConfig.ServerAddr, agentConfig.Key),
		agentConfig.PoolInterval,
		agentConfig.ReportInterval,
	)
	newAgent.Run()
}
