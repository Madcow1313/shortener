package config

import (
	"flag"
	"os"
)

type Config struct {
	Host        string
	BaseURL     string
	URLStorage  string
	DatabaseDSN string
}

const (
	defaultHost        = "localhost:8080"
	defaultBaseURL     = ""
	defaultURLstorage  = "/tmp/short-url-db.json"
	defaultDatabaseDSN = "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable"
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
	if storage := os.Getenv("FILE_STORAGE_PATH"); storage != "" {
		c.URLStorage = storage
	}
	if databaseDSN := os.Getenv("DATABASE_DSN"); databaseDSN != "" {
		c.DatabaseDSN = databaseDSN
	}
}
