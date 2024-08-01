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

func NewSimpleServer(conf config.Config, storage *os.File) *SimpleServer {
	return &SimpleServer{Host: conf.Host, BaseURL: conf.BaseURL, Storage: storage, URLmap: map[string]string{},
		UserURLS: map[string][]string{}, ID: 1, Config: conf, Compressor: &compressor.Compressor{}}
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
	err := c.ConnectToDB(func(db *sql.DB, args ...interface{}) error {
		return c.CreateTable(db)
	})
	if err == nil {
		err := c.ConnectToDB(func(db *sql.DB, args ...interface{}) error {
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

func (s *SimpleServer) RouteHandlers(router *chi.Mux, hh handlers.HandlerHelper) {
	ba := auth.NewBasicAuth(s.Config.SecretKey)

	router.HandleFunc("/", ba.CheckUserCookies(s.Compressor.Decompress(
		hh.ZapLogger.LogRequest(hh.HandlePostURL()))))

	router.HandleFunc("/api/shorten", ba.CheckUserCookies(s.Compressor.Decompress(
		hh.ZapLogger.LogRequest(hh.HandlePostAPIShorten()))))

	router.HandleFunc("/api/shorten/batch", ba.CheckUserCookies(s.Compressor.Decompress(
		hh.ZapLogger.LogRequest(hh.HandlePostAPIShortenBatch()))))

	router.HandleFunc("/ping", hh.HandlePing())

	router.Get("/api/user/urls", ba.AuthentificateUser(hh.ZapLogger.LogRequest(hh.HandleGetAPIUserURLs())))
	router.Delete("/api/user/urls", ba.AuthentificateUser(hh.ZapLogger.LogRequest(hh.HandleDeleteAPIUserURLs())))
}

func (s *SimpleServer) RunServer() {
	var hh handlers.HandlerHelper
	var mylogger mylogger.Mylogger
	var baseURL string
	router := chi.NewRouter()

	err := mylogger.Initialize("INFO")
	if err != nil {
		log.Fatal(err)
	}
	hh.Config = s.Config
	hh.Connector = dbconnector.NewConnector(hh.Config.DatabaseDSN)

	serv := server.SimpleServer{
		Host:    s.Host,
		BaseURL: s.BaseURL,
		URLmap:  s.URLmap,
		Storage: s.Storage,
		ID:      s.ID,
		Config:  s.Config,
	}
	serv.URLsToUpdate = make(chan string, 100)
	defer close(serv.URLsToUpdate)

	hh.Server = &serv
	hh.ZapLogger = mylogger
	hh.Connector.Z = &mylogger
	hh.Router = router
	hh.UserURLS = s.UserURLS

	if s.BaseURL != "" {
		baseURL = s.BaseURL + "/"
	}
	for k, v := range s.URLmap {
		router.Get("/"+baseURL+k, s.Compressor.Compress(mylogger.LogRequest(hh.HandleGetPostedURL("/"+k, v))))
	}

	s.RouteHandlers(router, hh)

	err = http.ListenAndServe(s.Host, router)
	if err != nil {
		mylogger.LogError(err)
	}
}
