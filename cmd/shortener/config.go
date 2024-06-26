package main

import (
	"flag"
	"os"
)

type config struct {
	Host    string
	BaseURL string
}

func (c *config) Set() {
	flag.StringVar(&c.Host, "a", "localhost:8080", "address should be in format localhost:8080")
	flag.StringVar(&c.BaseURL, "b", "", "base url should contain at least one character")
	flag.Parse()

	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		c.Host = addr
	}
	if base := os.Getenv("BASE_URL"); base != "" {
		c.BaseURL = base
	}
}
