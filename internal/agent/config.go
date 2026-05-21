package agent

import (
	"flag"
	"fmt"
	"regexp"

	"github.com/caarlos0/env/v6"
)

const (
	defaultServerAddr     = "http://localhost:8080"
	defaultPoolInterval   = int64(2)
	defaultReportInterval = int64(10)
	defaultKey            = ""
)

type Config struct {
	ServerAddr     string `env:"ADDRESS"        `
	PoolInterval   int64  `env:"POLL_INTERVAL"  `
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
}

var addrRegexp = regexp.MustCompile(`^(https?://)?(localhost|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(\d{2,5})$`)

func parseAddr(value string) (string, error) {
	matches := addrRegexp.FindStringSubmatch(value)
	if matches == nil {
		return "", fmt.Errorf("invalid address: %s", value)
	}
	scheme := matches[1]
	if scheme == "" {
		scheme = "http://"
	}
	return fmt.Sprintf("%s%s:%s", scheme, matches[2], matches[3]), nil
}

func GetConfig() (*Config, error) {
	var config Config
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	addrFlag := flag.String("a", defaultServerAddr, "server address host:port")
	poolIntervalFlag := flag.Int64("p", defaultPoolInterval, "pool interval in seconds")
	reportIntervalFlag := flag.Int64("r", defaultReportInterval, "report interval in seconds")
	keyFlag := flag.String("k", defaultKey, "key to calculate data hash")
	flag.Parse()

	if config.ServerAddr == "" {
		config.ServerAddr = *addrFlag
	}
	if config.PoolInterval == 0 {
		config.PoolInterval = *poolIntervalFlag
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = *reportIntervalFlag
	}
	if config.Key == "" {
		config.Key = *keyFlag
	}
	addr, err := parseAddr(config.ServerAddr)
	if err != nil {
		return nil, err
	}
	config.ServerAddr = addr

	return &config, nil
}
