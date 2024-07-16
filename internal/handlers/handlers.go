package handlers

import (
	"crypto/rand"
	"database/sql"
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

type DataBatchJSON struct {
	CorrelationID string `json:"correlation_id"`
	URL           string `json:"original_url"`
}

type BatchJSON struct {
	Data []DataBatchJSON
}

type ResponseJSON struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type HandlerHelper struct {
	Config config.Config
	Server server.SimpleServer
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

		err = hh.WriteToStorage(str, string(b))
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
		err = hh.WriteToStorage(str, string(b))
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

func (hh *HandlerHelper) WriteToStorage(shortened string, origin string) error {
	var err error
	switch hh.Config.StorageType {
	case config.Database:
		c := dbconnector.NewConnector(hh.Config.DatabaseDSN)
		err = c.Connect(func(db *sql.DB, args ...interface{}) error {
			return c.InsertURL(db, shortened, origin)
		})
	case config.File:
		err = WriteToFileStorage(hh.Server.Storage, origin, shortened, hh.Server.ID)

	default:
	}
	return err
}

func (hh *HandlerHelper) HandlePostAPIShortenBatch(s *server.SimpleServer, router *chi.Mux) http.HandlerFunc {
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

		var bJSON BatchJSON

		bJSON.Data = make([]DataBatchJSON, 0)
		err = json.Unmarshal(b, &bJSON.Data)
		if err != nil {
			http.Error(w, "Unable to parse json", http.StatusBadRequest)
			return
		}

		data := make(map[string]string)
		responseData := make(map[string]string)
		var baseURL string
		if s.BaseURL != "" {
			baseURL = s.BaseURL + "/"
		}
		for _, val := range bJSON.Data {
			shortURL := hh.ShortenURL()

			router.Get("/"+baseURL+shortURL, compressor.Compress(mylogger.LogRequest(hh.HandleGetPostedURL(s, "/"+shortURL, val.URL))))

			s.URLmap[shortURL] = val.URL
			data[shortURL] = val.URL
			responseData[shortURL] = val.CorrelationID

			if hh.Config.StorageType == config.File {
				err = WriteToFileStorage(s.Storage, val.URL, shortURL, hh.Server.ID)
				if err != nil {
					mylogger.LogError(err)
				} else {
					hh.Server.ID++
				}
			}
		}

		if hh.Config.StorageType == config.Database {
			c := dbconnector.NewConnector(hh.Config.DatabaseDSN)
			err := c.Connect(func(db *sql.DB, args ...interface{}) error {
				return c.InsertBatchToDatabase(db, data)
			})
			if err != nil {
				mylogger.LogError(err)
			}
		}

		temp := make([]ResponseJSON, 0)
		for key, value := range responseData {
			temp = append(temp, ResponseJSON{value, "http://" + s.Host + "/" + baseURL + key})
		}
		response, err := json.MarshalIndent(temp, "	", "")
		if err != nil {
			http.Error(w, "Unable to marshal response", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(response)), 10))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(response))
	}
}

func WriteToFileStorage(file *os.File, originalURL string, shortURL string, id int64) error {
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
