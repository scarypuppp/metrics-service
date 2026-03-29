package main

import (
	"net/http"

	"github.com/scarypuppp/metrics-service/internal/agent"
)

func main() {
	collector := agent.Collector{}
	sender := agent.Sender{
		Client:  &http.Client{},
		BaseURL: "http://127.0.0.1:8080",
	}

	newAgent := agent.Agent{
		Collector: &collector,
		Sender:    &sender,
	}
	newAgent.Run()
}
