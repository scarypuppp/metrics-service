package config

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
	defaultAddr            = "localhost:8080"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "metrics.json"
	defaultRestore         = true
)

// Config holds server configuration populated from environment variables and command-line flags.
type Config struct {
	Addr            string `env:"ADDRESS"        `
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	CryptoKey       string `env:"CRYPTO_KEY"`
	ConfigFile      string `env:"CONFIG"`
}

// ConfigJSON represents config from json.
type ConfigJSON struct {
	Addr          string `json:"address"`
	Restore       *bool  `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
}

// GetConfig builds server Config
func GetConfig() (*Config, error) {
	var config Config

	// 1. env
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	// 2. flags
	addrFlag := flag.String("a", "", "server address host:port")
	storeIntervalFlag := flag.Int("i", 0, "store metrics to file interval")
	fileStoragePathFlag := flag.String("f", "", "metrics storage file path")
	restoreFlag := flag.Bool("r", false, "restore metrics from file")
	databaseDsnFlag := flag.String("d", "", "database dsn string")
	keyFlag := flag.String("k", "", "key to calculate data hash")
	auditFileFlag := flag.String("audit-file", "", "audit file path")
	auditURLFlag := flag.String("audit-url", "", "audit url path")
	cryptoKeyFlag := flag.String("crypto-key", "", "private key path")
	configFileFlag := flag.String("c", "", "config file path")
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
	if config.Restore == nil && flagPassed("r") {
		config.Restore = restoreFlag
	}
	if config.DatabaseDSN == "" {
		config.DatabaseDSN = *databaseDsnFlag
	}
	if config.Key == "" {
		config.Key = *keyFlag
	}
	if config.AuditFile == "" {
		config.AuditFile = *auditFileFlag
	}
	if config.AuditURL == "" {
		config.AuditURL = *auditURLFlag
	}
	if config.CryptoKey == "" {
		config.CryptoKey = *cryptoKeyFlag
	}
	if config.ConfigFile == "" {
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

	addr, err := parseAddr(config.Addr)
	if err != nil {
		return nil, err
	}
	config.Addr = addr
	return &config, nil
}

// flagPassed reports whether the flag was actually set in the command line.
func flagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// setDefaults fills values that no source provided.
func setDefaults(config *Config) {
	if config.Addr == "" {
		config.Addr = defaultAddr
	}
	if config.StoreInterval == 0 {
		config.StoreInterval = defaultStoreInterval
	}
	if config.FileStoragePath == "" {
		config.FileStoragePath = defaultFileStoragePath
	}
	if config.Restore == nil {
		restore := defaultRestore
		config.Restore = &restore
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

	if config.Addr == "" {
		config.Addr = configJSON.Addr
	}
	if config.Restore == nil {
		config.Restore = configJSON.Restore
	}
	if config.StoreInterval == 0 && configJSON.StoreInterval != "" {
		dur, err := time.ParseDuration(configJSON.StoreInterval)
		if err != nil {
			return err
		}
		config.StoreInterval = int(dur.Seconds())
	}
	if config.FileStoragePath == "" {
		config.FileStoragePath = configJSON.StoreFile
	}
	if config.DatabaseDSN == "" {
		config.DatabaseDSN = configJSON.DatabaseDSN
	}
	if config.CryptoKey == "" {
		config.CryptoKey = configJSON.CryptoKey
	}
	return nil
}

var addrRegexp = regexp.MustCompile(`^(https?://)?(localhost|\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})?:(\d{2,5})$`)

func parseAddr(value string) (string, error) {
	matches := addrRegexp.FindStringSubmatch(value)
	if matches == nil {
		return "", fmt.Errorf("invalid address: %s", value)
	}
	return fmt.Sprintf("%s:%s", matches[2], matches[3]), nil
}
