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
	defaultRateLimit      = 1
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

// GetConfig builds agent Config.
func GetConfig() (*Config, error) {
	var config Config

	// 1. env
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	// 2. flags
	addrFlag := flag.String("a", "", "server address host:port")
	poolIntervalFlag := flag.Int64("p", 0, "pool interval in seconds")
	reportIntervalFlag := flag.Int64("r", 0, "report interval in seconds")
	keyFlag := flag.String("k", "", "key to calculate data hash")
	rateLimitFlag := flag.Int("l", 0, "rate limit to send data")
	cryptoKeyFlag := flag.String("crypto-key", "", "public key path")
	configFileFlag := flag.String("c", "", "config file path")
	flag.Parse()

	if config.ServerAddr == "" && flagPassed("a") {
		config.ServerAddr = *addrFlag
	}
	if config.PoolInterval == 0 && flagPassed("p") {
		config.PoolInterval = *poolIntervalFlag
	}
	if config.ReportInterval == 0 && flagPassed("r") {
		config.ReportInterval = *reportIntervalFlag
	}
	if config.Key == "" && flagPassed("k") {
		config.Key = *keyFlag
	}
	if config.RateLimit == 0 && flagPassed("l") {
		config.RateLimit = *rateLimitFlag
	}
	if config.CryptoKey == "" && flagPassed("crypto-key") {
		config.CryptoKey = *cryptoKeyFlag
	}
	if config.ConfigFile == "" && flagPassed("c") {
		config.ConfigFile = *configFileFlag
	}

	// 3. json
	if config.ConfigFile != "" {
		if err := configFromJSON(&config); err != nil {
			return nil, fmt.Errorf("error reading config from json: %w", err)
		}
	}

	// 4. defaults
	setDefaults(&config)

	addr, err := parseAddr(config.ServerAddr)
	if err != nil {
		return nil, err
	}
	config.ServerAddr = addr

	return &config, nil
}

// setDefaults fills values that no source provided.
func setDefaults(config *Config) {
	if config.ServerAddr == "" {
		config.ServerAddr = defaultServerAddr
	}
	if config.PoolInterval == 0 {
		config.PoolInterval = defaultPoolInterval
	}
	if config.ReportInterval == 0 {
		config.ReportInterval = defaultReportInterval
	}
	if config.RateLimit == 0 {
		config.RateLimit = defaultRateLimit
	}
}

// configFromJSON fills config empty values from json file.
func configFromJSON(config *Config) error {
	fileBytes, err := os.ReadFile(config.ConfigFile)
	if err != nil {
		return err
	}
	var configJSON ConfigJSON
	if err := json.Unmarshal(fileBytes, &configJSON); err != nil {
		return err
	}

	if config.ServerAddr == "" {
		config.ServerAddr = configJSON.ServerAddr
	}
	if config.PoolInterval == 0 && configJSON.PoolInterval != "" {
		dur, err := time.ParseDuration(configJSON.PoolInterval)
		if err != nil {
			return err
		}
		config.PoolInterval = int64(dur.Seconds())
	}
	if config.ReportInterval == 0 && configJSON.ReportInterval != "" {
		dur, err := time.ParseDuration(configJSON.ReportInterval)
		if err != nil {
			return err
		}
		config.ReportInterval = int64(dur.Seconds())
	}
	if config.CryptoKey == "" {
		config.CryptoKey = configJSON.CryptoKey
	}

	return nil
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

func flagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
