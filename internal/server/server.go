package server

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
)

type Server interface {
	RunServer()
}

type SimpleServer struct {
	Host,
	BaseURL string
	URLmap map[string]string
}

func shortenURL() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := make([]rune, 6)
	for i := 0; i < 6; i++ {
		index := rand.Intn(len(letters))
		result[i] = letters[index]
	}
	return string(result)
}

func InitServer(host, baseURL string) Server {
	return SimpleServer{Host: host, BaseURL: baseURL, URLmap: map[string]string{}}
}

func handleMainPage(s *SimpleServer, router *http.ServeMux) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		str := shortenURL()
		s.URLmap[str] = string(b)
		router.HandleFunc("/"+str, handleGetID(s, string(b)))
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		respBody := s.Host + "/" + s.BaseURL + str
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))
		w.Write([]byte(respBody))
	}
}

func handleGetID(s *SimpleServer, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", path)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (s SimpleServer) RunServer() {
	router := http.NewServeMux()

	router.HandleFunc("/", handleMainPage(&s, router))
	// router.HandleFunc("/{id}/", handleGetId(&s))

	err := http.ListenAndServe(s.Host, router)
	if err != nil {
		fmt.Println(fmt.Errorf("can't run server: %w", err))
	}
}
