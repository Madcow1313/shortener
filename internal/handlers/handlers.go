package handlers

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"os"
	"shortener/cmd/shortener/config"
	"shortener/internal/compressor"
	"shortener/internal/dbconnector"
	"shortener/internal/mylogger"
	server "shortener/internal/server/serverTypes"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type SimpleServer server.SimpleServer

type DataJSON struct {
	URL string `json:"url"`
}

type HandlerHelper struct {
	Config config.Config
}

func (hh *HandlerHelper) ShortenURL() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := make([]rune, 6)
	for i := 0; i < 6; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			mylogger.LogError(err)
			continue
		}
		result[i] = letters[index.Int64()]
	}
	return string(result)
}

func (hh *HandlerHelper) HandlePostURL(s *server.SimpleServer, router *chi.Mux) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Invalid request method", http.StatusBadRequest)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		str := hh.ShortenURL()
		s.URLmap[str] = string(b)
		err = WriteToStorage(s.Storage, string(b), str, s.ID)
		if err != nil {
			mylogger.LogError(err)
		} else {
			s.ID++
		}

		var baseURL string
		if s.BaseURL != "" {
			baseURL = s.BaseURL + "/"
		}
		router.Get("/"+baseURL+str, compressor.Compress(mylogger.LogRequest(hh.HandleGetPostedURL(s, "/"+str, string(b))))) //так и не смог придумать как убрать отсюда роутер

		w.Header().Set("Content-Type", "text/plain")
		respBody := "http://" + s.Host + "/" + baseURL + str
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(respBody))
	}
}

func (hh *HandlerHelper) HandleGetPostedURL(s *server.SimpleServer, path string, origin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Location", origin)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (hh *HandlerHelper) HandlePostAPIShorten(s *server.SimpleServer, router *chi.Mux) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Invalid request method", http.StatusBadRequest)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Unable to parse form", http.StatusBadRequest)
			return
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var d DataJSON
		err = json.Unmarshal(b, &d)
		if err != nil {
			http.Error(w, "Unable to unmarshal json-data", http.StatusBadRequest)
			return
		}

		str := hh.ShortenURL()
		err = WriteToStorage(s.Storage, string(b), str, s.ID)
		if err != nil {
			mylogger.LogError(err)
		} else {
			s.ID++
		}

		s.URLmap[str] = d.URL
		var baseURL string
		if s.BaseURL != "" {
			baseURL = s.BaseURL + "/"
		}
		router.Get("/"+baseURL+str, compressor.Compress(mylogger.LogRequest(hh.HandleGetPostedURL(s, "/"+str, d.URL))))

		w.Header().Set("Content-Type", "application/json")
		res := map[string]string{
			"result": "http://" + s.Host + "/" + baseURL + str,
		}
		respBody, err := json.MarshalIndent(res, "", "	")
		if err != nil {
			http.Error(w, "Unable to marshal response", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(respBody))
	}
}

func WriteToStorage(file *os.File, originalURL string, shortURL string, id int64) error {
	data := server.URLDataJSON{
		ID:          id,
		OriginalURL: originalURL,
		ShortURL:    shortURL,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = file.WriteString(string(b) + "\n")
	return err
}

func (hh *HandlerHelper) HandlePing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := dbconnector.NewConnector(hh.Config.DatabaseDSN)
		err := c.Connect(nil)
		if err != nil {
			http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
