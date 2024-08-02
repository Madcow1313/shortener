package config

import (
	"flag"
	"os"
)

const (
	Memory = iota
	File
	Database
)

type Config struct {
	Host        string
	BaseURL     string
	URLStorage  string
	DatabaseDSN string
	SecretKey   string
	StorageType int
}

const (
	defaultHost        = "localhost:8080"
	defaultBaseURL     = ""
	defaultURLstorage  = ""
	defaultDatabaseDSN = ""
	defaultSecretKey   = "cookieinthejar"

	hostHint       = "address should be in format localhost:8080"
	baseURLHint    = "base url should contain at least one character"
	URLStorageHint = "file to store urls"
	dbHint         = "database connection properties"
	keyHint        = "key to check hash sum"

	envServerAddress = "SERVER_ADDRESS"
	envBaseURL       = "BASE_URL"
	envDatabaseDSN   = "DATABASE_DSN"
	envFilePath      = "FILE_STORAGE_PATH"
	envKey           = "SECRET_KEY"
)

func (c *Config) SetConfigParameteres() {
	flag.StringVar(&c.Host, "a", defaultHost, hostHint)
	flag.StringVar(&c.BaseURL, "b", defaultBaseURL, baseURLHint)
	flag.StringVar(&c.URLStorage, "f", defaultURLstorage, URLStorageHint)
	flag.StringVar(&c.DatabaseDSN, "d", defaultDatabaseDSN, dbHint)
	flag.StringVar(&c.SecretKey, "s", defaultSecretKey, keyHint)
	flag.Parse()

	if addr := os.Getenv(envServerAddress); addr != "" {
		c.Host = addr
	}
	if base := os.Getenv(envBaseURL); base != "" {
		c.BaseURL = base
	}
	if databaseDSN := os.Getenv(envDatabaseDSN); databaseDSN != "" {
		c.DatabaseDSN = databaseDSN
	}
	if storage := os.Getenv(envFilePath); storage != "" {
		c.URLStorage = storage
	}
	if secretKey := os.Getenv(envKey); secretKey != "" {
		c.SecretKey = secretKey
	}

	c.StorageType = Memory
	if c.DatabaseDSN == "" {
		if c.URLStorage != "" {
			c.StorageType = File
		}
	} else {
		c.StorageType = Database
	}
}
