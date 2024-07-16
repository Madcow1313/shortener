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
	StorageType int
}

const (
	defaultHost        = "localhost:8080"
	defaultBaseURL     = ""
	defaultURLstorage  = ""
	defaultDatabaseDSN = "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable"
	// defaultDatabaseDSN = ""
)

func (c *Config) SetConfigParameteres() {
	flag.StringVar(&c.Host, "a", defaultHost, "address should be in format localhost:8080")
	flag.StringVar(&c.BaseURL, "b", defaultBaseURL, "base url should contain at least one character")
	flag.StringVar(&c.URLStorage, "f", defaultURLstorage, "file to store urls")
	flag.StringVar(&c.DatabaseDSN, "d", defaultDatabaseDSN, "database connection properties")
	flag.Parse()

	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		c.Host = addr
	}
	if base := os.Getenv("BASE_URL"); base != "" {
		c.BaseURL = base
	}
	if databaseDSN := os.Getenv("DATABASE_DSN"); databaseDSN != "" {
		c.DatabaseDSN = databaseDSN
	}
	if storage := os.Getenv("FILE_STORAGE_PATH"); storage != "" {
		c.URLStorage = storage
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
