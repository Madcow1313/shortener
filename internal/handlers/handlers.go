package handlers

import (
	"io"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

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

func HandleMainPage(s *SimpleServer, router *chi.Mux) http.HandlerFunc {
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
		var pathID string
		if s.BaseURL != "/" {
			pathID = s.BaseURL + "/"
		} else {
			pathID = s.BaseURL
		}
		router.Get(pathID+str, HandleGetID(s, "/"+str, string(b)))
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		respBody := "http://" + s.Host + pathID + str
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))
		w.Write([]byte(respBody))
	}
}

func HandleGetID(s *SimpleServer, path string, origin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", origin)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}
