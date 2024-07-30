package server

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"shortener/cmd/shortener/config"
	"shortener/internal/auth"
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
	return &SimpleServer{Host: conf.Host, BaseURL: conf.BaseURL, Storage: storage, URLmap: map[string]string{},
		UserURLS: map[string][]string{}, ID: 1, Config: conf}
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
			s.UserURLS = c.UserURLS
		}
	}
	c.DB.Close()
	return err
}

func (s *SimpleServer) RunServer() {
	router := chi.NewRouter()
	var hh handlers.HandlerHelper
	hh.Config = s.Config
	hh.Connector = dbconnector.NewConnector(hh.Config.DatabaseDSN)
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
	serv.URLsToUpdate = make(chan string)
	defer close(serv.URLsToUpdate)
	hh.Server = &serv
	hh.Z = mylogger
	hh.Connector.Z = &mylogger
	hh.Router = router
	hh.UserURLS = s.UserURLS

	var baseURL string
	ba := auth.NewBasicAuth(s.Config.SecretKey)
	if s.BaseURL != "" {
		baseURL = s.BaseURL + "/"
	}
	for k, v := range s.URLmap {
		router.Get("/"+baseURL+k, compressor.Compress(mylogger.LogRequest(hh.HandleGetPostedURL("/"+k, v))))
	}
	router.HandleFunc("/", ba.CheckCookies(compressor.Decompress(
		mylogger.LogRequest(hh.HandlePostURL()))))

	router.HandleFunc("/api/shorten", ba.CheckCookies(compressor.Decompress(
		mylogger.LogRequest(hh.HandlePostAPIShorten()))))

	router.HandleFunc("/ping", hh.HandlePing())

	router.HandleFunc("/api/shorten/batch", ba.CheckCookies(compressor.Decompress(
		mylogger.LogRequest(hh.HandlePostAPIShortenBatch()))))

	router.Get("/api/user/urls", ba.Authenticate(mylogger.LogRequest(hh.HandleGetAPIUserURLs())))

	router.Delete("/api/user/urls", ba.Authenticate(mylogger.LogRequest(hh.HandleDeleteAPIUserURLs())))

	err = http.ListenAndServe(s.Host, router)
	if err != nil {
		mylogger.LogError(err)
	}
}
