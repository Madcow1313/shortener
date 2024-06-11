package main

import (
	"flag"
)

type config struct {
	Host    string
	BaseURL string
}

func (c *config) Set() {
	flag.StringVar(&c.Host, "a", "localhost:8080", "address should be in format localhost:8080")
	flag.StringVar(&c.BaseURL, "b", "/", "base url should contain at least one character")
	flag.Parse()
}
