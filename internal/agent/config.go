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
)

type Config struct {
	ServerAddr     string `env:"ADDRESS"        `
	PoolInterval   int64  `env:"POLL_INTERVAL"  `
	ReportInterval int64  `env:"REPORT_INTERVAL"`
}

var addrRegexp = regexp.MustCompile(`^(https?://)?(localhost|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(\d{2,5})$`)

func parseAddr(value string) (string, error) {
	matches := addrRegexp.FindStringSubmatch(value)
	if matches == nil {
		return "", fmt.Errorf("invalid address: %s", value)
	}
	scheme := matches[1]
	fmt.Println("SCHEME", scheme)
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

	if config.ServerAddr == "" {
		flag.StringVar(&config.ServerAddr, "a", defaultServerAddr, "server address host:port")
	}
	if config.PoolInterval == 0 {
		flag.Int64Var(&config.PoolInterval, "p", defaultPoolInterval, "pool interval in seconds")
	}
	if config.ReportInterval == 0 {
		flag.Int64Var(&config.ReportInterval, "r", defaultReportInterval, "report interval in seconds")
	}
	flag.Parse()

	addr, err := parseAddr(config.ServerAddr)
	if err != nil {
		return nil, err
	}
	config.ServerAddr = addr

	return &config, nil
}
