package server

import (
	"os"
	"shortener/cmd/shortener/config"
	"shortener/internal/compressor"
	"shortener/internal/dbconnector"
)

type Server interface {
	RunServer()
}

type SimpleServer struct {
	ID int64
	Host,
	BaseURL string
	URLmap       map[string]string
	UserURLS     map[string][]string
	URLsToUpdate chan string
	Config       config.Config
	Storage      *os.File
	Connector    *dbconnector.Connector
	Compressor   *compressor.Compressor
}

type URLDataJSON struct {
	ID          int64  `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
