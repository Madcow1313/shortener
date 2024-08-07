package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"shortener/cmd/shortener/config"
	"strconv"
	"strings"
)

func (hh *HandlerHelper) HandleGetPostedURL(path string, origin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if hh.Config.StorageType == config.Database {
			temp := strings.Split(path, "/")
			short := temp[len(temp)-1]
			hh.Connector.ConnectToDB(func(db *sql.DB, args ...interface{}) error {
				return hh.Connector.IsShortDeleted(db, short)
			})
			if hh.Connector.IsDeleted {
				w.WriteHeader(http.StatusGone)
				return
			}
		}
		w.Header().Add("Location", origin)
		w.WriteHeader(http.StatusTemporaryRedirect)
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
			http.Error(w, marshalResponseError, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.FormatInt(int64(len(respBody)), 10))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(respBody))
	}
}

func (hh *HandlerHelper) HandlePing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := hh.Connector.ConnectToDB(nil)
		if err != nil {
			http.Error(w, pingError, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
