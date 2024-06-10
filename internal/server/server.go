package server

import (
	"fmt"
	"net/http"
	"shortener/internal/handlers"

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
	// router := http.NewServeMux()

	router := chi.NewRouter()
	router.HandleFunc("/", handlers.HandleMainPage(&handlers.SimpleServer{
		Host:    s.Host,
		BaseURL: s.BaseURL,
		URLmap:  s.URLmap,
	}, router))

	err := http.ListenAndServe(s.Host, router)
	if err != nil {
		fmt.Println(fmt.Errorf("can't run server: %w", err))
	}
}
