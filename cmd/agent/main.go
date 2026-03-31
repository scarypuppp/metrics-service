package main

import (
	"flag"
	"fmt"
	"net/http"
	"regexp"

	"github.com/scarypuppp/metrics-service/internal/agent"
)

var agentOptions struct {
	serverAddr     string
	poolInterval   int64
	reportInterval int64
}

func main() {
	agentOptions.serverAddr = "https://localhost:8080"
	flag.Func("a", "server address host:port", func(flagValue string) error {
		expr, err := regexp.Compile(`^(https?://)?(localhost|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(\d{2,5})$`)
		if err != nil {
			return err
		}
		matches := expr.FindStringSubmatch(flagValue)
		if matches == nil {
			return fmt.Errorf("invalid address: %s", flagValue)
		}
		scheme := matches[1]
		if scheme == "" {
			scheme = "http://"
		}
		host := matches[2]
		port := matches[3]
		resultAddress := fmt.Sprintf("%s%s:%s", scheme, host, port)
		agentOptions.serverAddr = resultAddress
		return nil
	})
	flag.Int64Var(&agentOptions.poolInterval, "p", 2, "pool interval in seconds")
	flag.Int64Var(&agentOptions.reportInterval, "r", 10, "report interval in seconds")
	flag.Parse()

	client := http.Client{}
	newAgent := agent.NewAgent(
		agent.NewCollector(),
		agent.NewSender(&client, agentOptions.serverAddr),
		agentOptions.poolInterval,
		agentOptions.reportInterval,
	)
	newAgent.Run()
}
