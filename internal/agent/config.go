package agent

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/caarlos0/env/v6"
)

const (
	defaultServerAddr     = "http://localhost:8080"
	defaultPoolInterval   = int64(2)
	defaultReportInterval = int64(10)
	defaultKey            = ""
	defaultRateLimit      = 1
	defaultCryptoKey      = ""
	defaultConfigFile     = ""
)

// Config holds agent configuration populated from environment variables and command-line flags.
type Config struct {
	ServerAddr     string `env:"ADDRESS"        `
	PoolInterval   int64  `env:"POLL_INTERVAL"  `
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	ConfigFile     string `env:"CONFIG"`
}

// ConfigJSON represents config from json.
type ConfigJSON struct {
	ServerAddr     string `json:"address"`
	PoolInterval   string `json:"poll_interval"`
	ReportInterval string `json:"report_interval"`
	CryptoKey      string `json:"crypto_key"`
}

// GetConfig builds agent Config from environment variables, falling back to command-line flags.
func GetConfig() (*Config, error) {
	var config Config
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	addrFlag := flag.String("a", defaultServerAddr, "server address host:port")
	poolIntervalFlag := flag.Int64("p", defaultPoolInterval, "pool interval in seconds")
	reportIntervalFlag := flag.Int64("r", defaultReportInterval, "report interval in seconds")
	keyFlag := flag.String("k", defaultKey, "key to calculate data hash")
	ratelimitKey := flag.Int("l", defaultRateLimit, "rate limit to send data")
	cryptoKeyFlag := flag.String("crypto-key", defaultCryptoKey, "public key path")
	configFileFlag := flag.String("c", defaultConfigFile, "config file path")

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
	if config.RateLimit == 0 {
		config.RateLimit = *ratelimitKey
	}
	if config.CryptoKey == "" {
		config.CryptoKey = *cryptoKeyFlag
	}
	if config.ConfigFile == "" {
		config.ConfigFile = *configFileFlag
	}

	if config.ConfigFile != "" {
		err := configFromJSON(&config)
		if err != nil {
			return nil, fmt.Errorf("error reading config from json: %w", err)
		}
	}

	addr, err := parseAddr(config.ServerAddr)
	if err != nil {
		return nil, err
	}
	config.ServerAddr = addr

	return &config, nil
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

// configFromJSON fills config empty values from
func configFromJSON(config *Config) error {
	fileBytes, err := os.ReadFile(config.ConfigFile)
	if err != nil {
		return err
	}
	var configJSON ConfigJSON
	err = json.Unmarshal(fileBytes, &configJSON)
	if err != nil {
		return err
	}

	if config.ServerAddr == "" && configJSON.ServerAddr != "" {
		config.ServerAddr = configJSON.ServerAddr
	}
	if config.PoolInterval == 0 && configJSON.PoolInterval != "" {
		dur, err := time.ParseDuration(configJSON.PoolInterval)
		if err != nil {
			return err
		}
		config.PoolInterval = int64(dur)
	}
	if config.ReportInterval == 0 && configJSON.ReportInterval != "" {
		dur, err := time.ParseDuration(configJSON.ReportInterval)
		if err != nil {
			return err
		}
		config.ReportInterval = int64(dur)
	}
	if config.CryptoKey == "" && configJSON.CryptoKey != "" {
		config.CryptoKey = configJSON.CryptoKey
	}

	return nil
}
