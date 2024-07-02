package server

import (
	"fmt"
	"net/http"
	"shortener/internal/compressor"
	"shortener/internal/handlers"
	"shortener/internal/mylogger"

	"github.com/go-chi/chi/v5"
)

type Server interface {
	RunServer()
}

type SimpleServer struct {
	Host,
	BaseURL string
	URLmap map[string]string
}

func InitServer(host, baseURL string) Server {
	return SimpleServer{Host: host, BaseURL: baseURL, URLmap: map[string]string{}}
}

func (s SimpleServer) RunServer() {
	router := chi.NewRouter()
	mylogger.Initialize("INFO")
	router.HandleFunc("/", compressor.Decompress(
		mylogger.LogRequest(handlers.HandleMainPage(&handlers.SimpleServer{
			Host:    s.Host,
			BaseURL: s.BaseURL,
			URLmap:  s.URLmap,
		}, router))))

	router.HandleFunc("/api/shorten", compressor.Decompress(
		mylogger.LogRequest(handlers.HandleAPIShorten(&handlers.SimpleServer{
			Host:    s.Host,
			BaseURL: s.BaseURL,
			URLmap:  s.URLmap,
		}, router))))

	err := http.ListenAndServe(s.Host, router)
	if err != nil {
		fmt.Println(fmt.Errorf("can't run server: %w", err))
	}
}
