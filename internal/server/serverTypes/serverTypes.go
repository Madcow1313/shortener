package server

import (
	"os"
	"shortener/cmd/shortener/config"
	"shortener/internal/dbconnector"
)

type Server interface {
	RunServer()
}

type SimpleServer struct {
	Host,
	BaseURL string
	URLmap    map[string]string
	ID        int64
	Storage   *os.File
	Config    config.Config
	Connector *dbconnector.Connector
}

type URLDataJSON struct {
	ID          int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
