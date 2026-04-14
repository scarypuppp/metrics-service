package config

import (
	"flag"
	"fmt"
	"regexp"

	"github.com/caarlos0/env/v6"
)

const (
	defaultAddr = "localhost:8080"
)

type Config struct {
	Addr string `env:"ADDRESS"        `
}

var addrRegexp = regexp.MustCompile(`^(https?://)?(localhost|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})?:(\d{2,5})$`)

func parseAddr(value string) (string, error) {
	matches := addrRegexp.FindStringSubmatch(value)
	if matches == nil {
		return "", fmt.Errorf("invalid address: %s", value)
	}
	host := matches[2]
	port := matches[3]
	return fmt.Sprintf("%s:%s", host, port), nil
}

func GetConfig() (*Config, error) {
	var config Config
	if err := env.Parse(&config); err != nil {
		return nil, err
	}
	if config.Addr == "" {
		flag.StringVar(&config.Addr, "p", defaultAddr, "server address host:port")
	}
	flag.Parse()

	addr, err := parseAddr(config.Addr)
	if err != nil {
		return nil, err
	}
	config.Addr = addr
	return &config, nil
}
