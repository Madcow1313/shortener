package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"shortener/internal/compressor"
	"shortener/internal/handlers"
	"shortener/internal/mylogger"
	server "shortener/internal/server/serverTypes"

	"github.com/go-chi/chi/v5"
)

type SimpleServer server.SimpleServer
type Server server.Server

func InitServer(host, baseURL string, storage *os.File) Server {
	return &SimpleServer{Host: host, BaseURL: baseURL, Storage: storage, URLmap: map[string]string{}, ID: 1}
}

func (s *SimpleServer) CheckStorage() error {
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

func (s *SimpleServer) RunServer() {
	router := chi.NewRouter()
	mylogger.Initialize("INFO")
	err := s.CheckStorage()
	if err != nil {
		fmt.Println(fmt.Errorf("can't read saved urls: %w", err))
	}
	serv := server.SimpleServer{
		Host:    s.Host,
		BaseURL: s.BaseURL,
		URLmap:  s.URLmap,
		Storage: s.Storage,
		ID:      s.ID,
	}

	var baseURL string
	if s.BaseURL != "" {
		baseURL = s.BaseURL + "/"
	}
	for k, v := range s.URLmap {
		router.Get("/"+baseURL+k, compressor.Compress(mylogger.LogRequest(handlers.HandleGetID(&serv, "/"+k, v))))
	}
	router.HandleFunc("/", compressor.Decompress(
		mylogger.LogRequest(handlers.HandleMainPage(&serv, router))))

	router.HandleFunc("/api/shorten", compressor.Decompress(
		mylogger.LogRequest(handlers.HandleAPIShorten(&serv, router))))

	err = http.ListenAndServe(s.Host, router)
	if err != nil {
		fmt.Println(fmt.Errorf("can't run server: %w", err))
	}
}
