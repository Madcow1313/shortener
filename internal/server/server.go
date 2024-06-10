package server

import (
	"fmt"
	"net/http"

	handlers "shortener/internal/handlers"
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
	router := http.NewServeMux()

	router.HandleFunc("/", handlers.HandleMainPage(&handlers.SimpleServer{
		Host:    s.Host,
		BaseURL: s.BaseURL,
		URLmap:  s.URLmap,
	}, router))
	// router.HandleFunc("/{id}/", handleGetId(&s))

	err := http.ListenAndServe(s.Host, router)
	if err != nil {
		fmt.Println(fmt.Errorf("can't run server: %w", err))
	}
}
