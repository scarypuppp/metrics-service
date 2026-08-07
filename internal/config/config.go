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
	defaultDatabaseDsn     = ""
	defaultKey             = ""
	defaultAuditFile       = ""
	defaultAuditURL        = ""
	defaultCryptoKey       = ""
	defaultConfigFile      = ""
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
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
}

// GetConfig builds server Config from environment variables, falling back to command-line flags.
func GetConfig() (*Config, error) {
	var config Config
	if err := env.Parse(&config); err != nil {
		return nil, err
	}

	addrFlag := flag.String("a", defaultAddr, "server address host:port")
	storeIntervalFlag := flag.Int("i", defaultStoreInterval, "store metrics to file interval")
	fileStoragePathFlag := flag.String("f", defaultFileStoragePath, "metrics storage file path")
	restoreFlag := flag.Bool("r", defaultRestore, "restore metrics from file")
	databaseDsnFlag := flag.String("d", defaultDatabaseDsn, "database dsn string")
	keyFlag := flag.String("k", defaultKey, "key to calculate data hash")
	auditFileFlag := flag.String("audit-file", defaultAuditFile, "audit file path")
	auditURLFlag := flag.String("audit-url", defaultAuditURL, "audit url path")
	cryptoKeyFlag := flag.String("crypto-key", defaultCryptoKey, "private key path")
	configFileFlag := flag.String("c", defaultConfigFile, "config file path")
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

	if config.ConfigFile != "" {
		err := configFromJSON(&config)
		if err != nil {
			return nil, fmt.Errorf("error reading config from json: %w", err)
		}
	}

	addr, err := parseAddr(config.Addr)
	if err != nil {
		return nil, err
	}
	config.Addr = addr
	return &config, nil
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

	if config.Addr == "" && configJSON.Addr != "" {
		config.Addr = configJSON.Addr
	}
	if config.Restore == nil {
		config.Addr = configJSON.Addr
	}
	if config.StoreInterval == 0 && configJSON.StoreInterval != "" {
		dur, err := time.ParseDuration(configJSON.StoreInterval)
		if err != nil {
			return err
		}
		config.StoreInterval = int(dur)
	}
	if config.FileStoragePath == "" && configJSON.StoreFile != "" {
		config.FileStoragePath = configJSON.StoreFile
	}
	if config.DatabaseDSN == "" && configJSON.DatabaseDSN != "" {
		config.DatabaseDSN = configJSON.DatabaseDSN
	}
	if config.CryptoKey == "" && configJSON.CryptoKey != "" {
		config.CryptoKey = configJSON.CryptoKey
	}
	return nil
}
