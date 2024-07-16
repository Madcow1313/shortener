package server

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"shortener/cmd/shortener/config"
	"shortener/internal/compressor"
	"shortener/internal/dbconnector"
	"shortener/internal/handlers"
	"shortener/internal/mylogger"
	server "shortener/internal/server/serverTypes"

	"github.com/go-chi/chi/v5"
)

type SimpleServer server.SimpleServer
type Server server.Server

func NewSimpleServer(conf config.Config, storage *os.File) *SimpleServer {
	return &SimpleServer{Host: conf.Host, BaseURL: conf.BaseURL, Storage: storage, URLmap: map[string]string{}, ID: 1, Config: conf}
}

func (s *SimpleServer) CheckFileStorage() error {
	scanner := bufio.NewScanner(s.Storage)
	id := 0
	for scanner.Scan() {
		var temp server.URLDataJSON
		err := json.Unmarshal([]byte(scanner.Text()), &temp)
		if err != nil {
			continue
		}
		id++
		s.URLmap[temp.ShortURL] = temp.OriginalURL
	}
	s.ID = int64(id + 1)
	return nil
}

func (s *SimpleServer) CheckDBStorage() error {
	c := dbconnector.NewConnector(s.Config.DatabaseDSN)
	err := c.Connect(func(db *sql.DB, args ...interface{}) error {
		return c.CreateTable(db)
	})
	if err == nil {
		err := c.Connect(func(db *sql.DB, args ...interface{}) error {
			return c.ReadFromDB(db)
		})
		if err == nil {
			s.URLmap = c.URLmap
		}
	}
	s.Connector = c
	return err
}

func (s *SimpleServer) RunServer() {
	router := chi.NewRouter()
	var hh handlers.HandlerHelper
	hh.Config = s.Config
	hh.Connector = s.Connector
	var mylogger mylogger.Mylogger
	err := mylogger.Initialize("INFO")
	if err != nil {
		log.Fatal(err)
	}
	serv := server.SimpleServer{
		Host:    s.Host,
		BaseURL: s.BaseURL,
		URLmap:  s.URLmap,
		Storage: s.Storage,
		ID:      s.ID,
		Config:  s.Config,
	}
	hh.Server = &serv
	hh.Z = mylogger
	hh.Router = router

	var baseURL string
	if s.BaseURL != "" {
		baseURL = s.BaseURL + "/"
	}
	for k, v := range s.URLmap {
		router.Get("/"+baseURL+k, compressor.Compress(mylogger.LogRequest(hh.HandleGetPostedURL("/"+k, v))))
	}
	router.HandleFunc("/", compressor.Decompress(
		mylogger.LogRequest(hh.HandlePostURL())))

	router.HandleFunc("/api/shorten", compressor.Decompress(
		mylogger.LogRequest(hh.HandlePostAPIShorten())))

	router.HandleFunc("/ping", hh.HandlePing())

	router.HandleFunc("/api/shorten/batch", compressor.Decompress(
		mylogger.LogRequest(hh.HandlePostAPIShortenBatch())))

	err = http.ListenAndServe(s.Host, router)
	if err != nil {
		mylogger.LogError(err)
	}
}
