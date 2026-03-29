package main

import (
	"net/http"

	"github.com/scarypuppp/metrics-service/internal/agent"
)

const poolInterval int64 = 2
const reportInterval int64 = 10

func main() {
	client := http.Client{}
	newAgent := agent.NewAgent(
		agent.NewCollector(),
		agent.NewSender(&client, "http://127.0.0.1:8080"),
		poolInterval,
		reportInterval,
	)
	newAgent.Run()
}
