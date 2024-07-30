package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
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
	Config    config.Config
	Server    *server.SimpleServer
	Connector *dbconnector.Connector
	Z         mylogger.Mylogger
	Router    *chi.Mux
	UserURLS  map[string][]string
}

func (hh *HandlerHelper) GetUserIDFromCookie(w http.ResponseWriter, r *http.Request) string {
	cookie := w.Header().Get("Set-Cookie")
	if cookie == "" {
		cookies, err := r.Cookie("user_id")
		if err != nil {
			return ""
		}
		cookie = cookies.Value
	}
	temp := strings.Split(cookie, " ")
	userID, _ := strings.CutPrefix(temp[0], "user_id=")
	return userID
}

func (hh *HandlerHelper) AddUserURL(userID string, url string) {
	if _, ok := hh.UserURLS[userID]; !ok {
		hh.UserURLS[userID] = make([]string, 0)
	}
	hh.UserURLS[userID] = append(hh.UserURLS[userID], url)
}

func (hh *HandlerHelper) ShortenURL() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	result := make([]rune, 6)
	for i := 0; i < 6; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			hh.Z.LogError(err)
			continue
		}
		result[i] = letters[index.Int64()]
	}
	return string(result)
}

func (hh *HandlerHelper) HandlePostURL() http.HandlerFunc {
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

		shortURL := hh.ShortenURL()
		var baseURL string
		inDatabase := false
		if hh.Server.BaseURL != "" {
			baseURL = hh.Server.BaseURL + "/"
		}
		userID := hh.GetUserIDFromCookie(w, r)
		err = hh.WriteToStorage(shortURL, string(b), userID)
		hh.AddUserURL(userID, shortURL)
		if err != nil {
			var pqErr *pq.Error
			if hh.Config.StorageType == config.Database && errors.As(err, &pqErr) {
				if pqErr.Code == pgerrcode.UniqueViolation {
					hh.Z.LogError(err)
					err = hh.Connector.Connect(func(db *sql.DB, args ...interface{}) error {
						return hh.Connector.SelectShortURL(db, string(b))
					})
					if err != nil {
						http.Error(w, "Unable to get short url from database", http.StatusInternalServerError)
						return
					}
					shortURL = hh.Connector.LastResult
					inDatabase = true
				}

			} else {
				hh.Z.LogError(err)
				http.Error(w, fmt.Errorf("error while working with database: %w", err).Error(), http.StatusInternalServerError)
				return
			}
		} else {
			hh.Server.ID++
			hh.Server.URLmap[shortURL] = string(b)
			hh.Router.Get("/"+baseURL+shortURL, compressor.Compress(hh.Z.LogRequest(hh.HandleGetPostedURL("/"+shortURL, string(b))))) //нет возможности (или я её не вижу), чтобы вынести роутинг отдельно от обработчика, ведь url для сокращения мы получаем из запроса

		}
		w.Header().Set("Content-Type", "text/plain")
		respBody := "http://" + hh.Server.Host + "/" + baseURL + shortURL
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))
		if !inDatabase {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusConflict)
		}
		w.Write([]byte(respBody))
	}
}

func (hh *HandlerHelper) HandleGetPostedURL(path string, origin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if hh.Config.StorageType == config.Database {
			temp := strings.Split(path, "/")
			short := temp[len(temp)-1]
			err := hh.Connector.Connect(func(db *sql.DB, args ...interface{}) error {
				return hh.Connector.IsShortDeleted(db, short)
			})
			if err != nil {
				http.Error(w, "No such url", http.StatusNotFound)
				return
			} else if hh.Connector.IsDeleted {
				w.WriteHeader(http.StatusGone)
				return
			}
		}
		w.Header().Add("Location", origin)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (hh *HandlerHelper) HandlePostAPIShorten() http.HandlerFunc {
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
		w.Header().Set("Content-Type", "application/json")

		shortURL := hh.ShortenURL()
		var baseURL string
		inDatabase := false
		if hh.Server.BaseURL != "" {
			baseURL = hh.Server.BaseURL + "/"
		}
		userID := hh.GetUserIDFromCookie(w, r)
		err = hh.WriteToStorage(shortURL, string(b), userID)
		hh.AddUserURL(userID, shortURL)
		if err != nil {
			var pqErr *pq.Error
			if hh.Config.StorageType == config.Database && errors.As(err, &pqErr) {
				if pqErr.Code == pgerrcode.UniqueViolation {
					hh.Z.LogError(err)
					err = hh.Connector.Connect(func(db *sql.DB, args ...interface{}) error {
						return hh.Connector.SelectShortURL(db, string(b))
					})
					if err != nil {
						http.Error(w, "Unable to get short url from database", http.StatusInternalServerError)
						return
					}
					shortURL = hh.Connector.LastResult
					inDatabase = true
				}

			} else {
				hh.Z.LogError(err)
				http.Error(w, fmt.Errorf("error while working with database: %w", err).Error(), http.StatusInternalServerError)
				return
			}
		} else {
			hh.Server.ID++
			hh.Server.URLmap[shortURL] = d.URL
			hh.Router.Get("/"+baseURL+shortURL, compressor.Compress(hh.Z.LogRequest(hh.HandleGetPostedURL("/"+shortURL, d.URL))))
		}

		res := map[string]string{
			"result": "http://" + hh.Server.Host + "/" + baseURL + shortURL,
		}
		respBody, err := json.MarshalIndent(res, "", "	")
		if err != nil {
			http.Error(w, "Unable to marshal response", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))
		if !inDatabase {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusConflict)
		}
		w.Write([]byte(respBody))
	}
}

func (hh *HandlerHelper) HandlePing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := hh.Connector.Connect(nil)
		if err != nil {
			http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (hh *HandlerHelper) WriteToStorage(shortened string, origin string, userID string) error {
	var err error
	switch hh.Config.StorageType {
	case config.Database:
		err = hh.Connector.Connect(func(db *sql.DB, args ...interface{}) error {
			return hh.Connector.InsertURL(db, shortened, origin, userID)
		})
	case config.File:
		err = WriteToFileStorage(hh.Server.Storage, origin, shortened, hh.Server.ID)

	default:
	}
	return err
}

func (hh *HandlerHelper) HandlePostAPIShortenBatch() http.HandlerFunc {
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
		if err != nil && err != io.EOF {
			http.Error(w, "Unable to parse json", http.StatusBadRequest)
			return
		}

		data := make(map[string]string)
		responseData := make(map[string]string)
		var baseURL string
		if hh.Server.BaseURL != "" {
			baseURL = hh.Server.BaseURL + "/"
		}
		for _, val := range bJSON.Data {
			shortURL := hh.ShortenURL()

			hh.Router.Get("/"+baseURL+shortURL, compressor.Compress(hh.Z.LogRequest(hh.HandleGetPostedURL("/"+shortURL, val.URL))))

			hh.Server.URLmap[shortURL] = val.URL
			data[shortURL] = val.URL
			responseData[shortURL] = val.CorrelationID
			hh.AddUserURL(hh.GetUserIDFromCookie(w, r), shortURL)
			if hh.Config.StorageType == config.File {
				err = WriteToFileStorage(hh.Server.Storage, val.URL, shortURL, hh.Server.ID)
				if err != nil {
					hh.Z.LogError(err)
				} else {
					hh.Server.ID++
				}
			}
		}

		if hh.Config.StorageType == config.Database {
			err := hh.Connector.Connect(func(db *sql.DB, args ...interface{}) error {
				return hh.Connector.InsertBatchToDatabase(db, data, hh.GetUserIDFromCookie(w, r))
			})
			if err != nil {
				hh.Z.LogError(err)
			}
		}

		temp := make([]ResponseJSON, 0)
		for key, value := range responseData {
			temp = append(temp, ResponseJSON{value, "http://" + hh.Server.Host + "/" + baseURL + key})
		}
		respBody, err := json.Marshal(temp)
		if err != nil {
			http.Error(w, "Unable to marshal response", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(respBody))
	}
}

func (hh *HandlerHelper) HandleGetAPIUserURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := hh.GetUserIDFromCookie(w, r)
		if _, ok := hh.UserURLS[userID]; !ok {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		type data struct {
			Short    string `json:"short_url"`
			Original string `json:"original_url"`
		}
		type allData struct {
			AllData []data
		}
		var ad allData
		ad.AllData = make([]data, 0)
		var baseURL string
		if hh.Server.BaseURL != "" {
			baseURL = hh.Server.BaseURL + "/"
		}
		for _, val := range hh.UserURLS[userID] {
			original := hh.Server.URLmap[val]
			d := data{
				Original: original,
				Short:    "http://" + hh.Server.Host + "/" + baseURL + val,
			}
			ad.AllData = append(ad.AllData, d)
		}
		respBody, err := json.MarshalIndent(ad.AllData, "", "	")
		if err != nil {
			http.Error(w, "Unable to marshal response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(respBody))
	}
}

func (hh *HandlerHelper) HandleDeleteAPIUserURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		urls := make([]string, 0)
		err = json.Unmarshal(b, &urls)
		if err != nil {
			http.Error(w, "urls cannot be unmarshalled", http.StatusBadRequest)
			return
		}
		// for i := 0; i < 1000; i++ {
		// 	str := strconv.FormatInt(int64(i), 10)
		// 	urls = append(urls, "abc"+str)
		// }
		// for i, val := range urls {
		// 	hh.Connector.Connect(func(db *sql.DB, args ...interface{}) error {
		// 		s := strconv.FormatInt(int64(i), 10)
		// 		return hh.Connector.InsertURL(db, val, s, "bc")
		// 	})
		// }

		ch := make(chan string, 100)
		go func() {
			for _, val := range urls {
				ch <- val
			}
			// close(ch)
		}()
		go func() {
			hh.Connector.Connect(func(db *sql.DB, args ...interface{}) error {
				return hh.Connector.UpdateOnDelete(db, hh.GetUserIDFromCookie(w, r), ch)
			})
		}()
		w.WriteHeader(http.StatusAccepted)
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
