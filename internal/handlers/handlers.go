package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"shortener/internal/compressor"
	"shortener/internal/mylogger"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type SimpleServer struct {
	Host,
	BaseURL string
	URLmap map[string]string
}

type DataJSON struct {
	URL string `json:"url"`
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
		var baseURL string
		if s.BaseURL != "" {
			baseURL = s.BaseURL + "/"
		}
		router.Get("/"+baseURL+str, compressor.Compress(mylogger.LogRequest(HandleGetID(s, "/"+str, string(b)))))
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		respBody := "http://" + s.Host + "/" + baseURL + str
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

func HandleAPIShorten(s *SimpleServer, router *chi.Mux) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err := r.ParseForm()
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var d DataJSON
		err = json.Unmarshal(b, &d)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		str := shortenURL()
		s.URLmap[str] = d.URL
		var baseURL string
		if s.BaseURL != "" {
			baseURL = s.BaseURL + "/"
		}
		router.Get("/"+baseURL+str, compressor.Compress(mylogger.LogRequest(HandleGetID(s, "/"+str, d.URL))))
		w.Header().Set("Content-Type", "application/json")

		res := map[string]string{
			"result": "http://" + s.Host + "/" + baseURL + str,
		}
		respBody, err := json.MarshalIndent(res, "", "	")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// respBody := "http://" + s.Host + "/" + baseURL + string(res)
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(respBody))
	}
}
