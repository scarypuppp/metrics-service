package config

import (
	"flag"
	"fmt"
	"regexp"

	"github.com/caarlos0/env/v6"
)

const (
	defaultAddr            = "localhost:8080"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "metrics.json"
	defaultRestore         = true
)

type Config struct {
	Addr            string `env:"ADDRESS"        `
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
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

	addrFlag := flag.String("a", defaultAddr, "server address host:port")
	storeIntervalFlag := flag.Int("i", defaultStoreInterval, "store metrics to file interval")
	fileStoragePathFlag := flag.String("f", defaultFileStoragePath, "metrics storage file path")
	restoreFlag := flag.Bool("r", defaultRestore, "restore metrics from file")
	flag.Parse()

	if config.Addr == "" {
		config.Addr = *addrFlag
	}
	if config.StoreInterval == 0 {
		config.StoreInterval = *storeIntervalFlag
	}
	if config.FileStoragePath == "" {
		config.FileStoragePath = *fileStoragePathFlag
	}
	if config.Restore == nil {
		config.Restore = restoreFlag
	}

	addr, err := parseAddr(config.Addr)
	if err != nil {
		return nil, err
	}
	config.Addr = addr
	return &config, nil
}
