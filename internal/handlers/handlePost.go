package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"shortener/cmd/shortener/config"
	"strconv"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

func (hh *HandlerHelper) HandlePostURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, invalidRequestError, http.StatusBadRequest)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, parseFormError, http.StatusBadRequest)
			return
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, bodyReadError, http.StatusBadRequest)
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
		var pqErr *pq.Error
		if err != nil && errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {
			hh.ZapLogger.LogError(err)
			hh.Connector.ConnectToDB(func(db *sql.DB, args ...interface{}) error {
				return hh.Connector.SelectShortURL(db, string(b))
			})
			if err != nil {
				http.Error(w, selectShortError, http.StatusInternalServerError)
				return
			}
			shortURL = hh.Connector.LastResult
			inDatabase = true
		} else if err != nil {
			hh.ZapLogger.LogError(err)
			http.Error(w, fmt.Errorf("error while working with database: %w", err).Error(), http.StatusInternalServerError)
			return
		} else {
			hh.Server.ID++
			hh.Server.URLmap[shortURL] = string(b)
			hh.Router.Get("/"+baseURL+shortURL, hh.Server.Compressor.Compress(hh.ZapLogger.LogRequest(hh.HandleGetPostedURL("/"+shortURL, string(b))))) //нет возможности (или я её не вижу), чтобы вынести роутинг отдельно от обработчика, ведь url для сокращения мы получаем из запроса
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

func (hh *HandlerHelper) HandlePostAPIShorten() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, invalidRequestError, http.StatusBadRequest)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, parseFormError, http.StatusBadRequest)
			return
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, bodyReadError, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var d DataJSON
		err = json.Unmarshal(b, &d)
		if err != nil {
			http.Error(w, unmarshalError, http.StatusBadRequest)
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
		var pqErr *pq.Error
		if err != nil && errors.As(err, &pqErr) && pqErr.Code == pgerrcode.UniqueViolation {

			hh.ZapLogger.LogError(err)
			err = hh.Connector.ConnectToDB(func(db *sql.DB, args ...interface{}) error {
				return hh.Connector.SelectShortURL(db, string(b))
			})
			if err != nil {
				http.Error(w, selectShortError, http.StatusInternalServerError)
				return
			}
			shortURL = hh.Connector.LastResult
			inDatabase = true

		} else if err != nil {
			hh.ZapLogger.LogError(err)
			http.Error(w, fmt.Errorf("error while working with database: %w", err).Error(), http.StatusInternalServerError)
			return
		} else {
			hh.Server.ID++
			hh.Server.URLmap[shortURL] = d.URL
			hh.Router.Get("/"+baseURL+shortURL, hh.Server.Compressor.Compress(hh.ZapLogger.LogRequest(hh.HandleGetPostedURL("/"+shortURL, d.URL))))
		}

		res := map[string]string{
			"result": "http://" + hh.Server.Host + "/" + baseURL + shortURL,
		}
		respBody, err := json.MarshalIndent(res, "", "	")
		if err != nil {
			http.Error(w, marshalResponseError, http.StatusBadRequest)
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

func (hh *HandlerHelper) HandlePostAPIShortenBatch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, invalidRequestError, http.StatusBadRequest)
			return
		}

		err := r.ParseForm()
		if err != nil {
			http.Error(w, parseFormError, http.StatusBadRequest)
			return
		}

		b, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, bodyReadError, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var bJSON BatchJSON

		bJSON.Data = make([]DataBatchJSON, 0)
		err = json.Unmarshal(b, &bJSON.Data)
		if err != nil && err != io.EOF {
			http.Error(w, unmarshalError, http.StatusBadRequest)
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

			hh.Router.Get("/"+baseURL+shortURL, hh.Server.Compressor.Compress(hh.ZapLogger.LogRequest(hh.HandleGetPostedURL("/"+shortURL, val.URL))))

			hh.Server.URLmap[shortURL] = val.URL
			data[shortURL] = val.URL
			responseData[shortURL] = val.CorrelationID
			hh.AddUserURL(hh.GetUserIDFromCookie(w, r), shortURL)
			if hh.Config.StorageType == config.File {
				err = WriteToFileStorage(hh.Server.Storage, val.URL, shortURL, hh.Server.ID)
				if err != nil {
					hh.ZapLogger.LogError(err)
				} else {
					hh.Server.ID++
				}
			}
		}

		if hh.Config.StorageType == config.Database {
			err := hh.Connector.ConnectToDB(func(db *sql.DB, args ...interface{}) error {
				return hh.Connector.InsertBatchToDatabase(db, data, hh.GetUserIDFromCookie(w, r))
			})
			if err != nil {
				hh.ZapLogger.LogError(err)
			}
		}

		temp := make([]ResponseJSON, 0)
		for key, value := range responseData {
			temp = append(temp, ResponseJSON{value, "http://" + hh.Server.Host + "/" + baseURL + key})
		}
		respBody, err := json.Marshal(temp)
		if err != nil {
			http.Error(w, marshalResponseError, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(respBody))
	}
}
