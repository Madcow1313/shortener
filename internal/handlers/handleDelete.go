package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

func (hh *HandlerHelper) HandleDeleteAPIUserURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		urls := make([]string, 0)
		err = json.Unmarshal(b, &urls)
		if err != nil {
			http.Error(w, unmarshalError, http.StatusBadRequest)
			return
		}
		ch := make(chan string, len(urls))
		go func() {
			for _, val := range urls {
				ch <- val
			}
			close(ch)
		}()
		go func() {
			ctxChild, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			select {
			case <-ctxChild.Done():
				return
			default:
				err = hh.Connector.ConnectToDB(func(db *sql.DB, args ...interface{}) error {
					return hh.Connector.UpdateIsDeletedColumn(db, ch)
				})
				if err != nil {
					hh.ZapLogger.LogError(err)
					return
				}
			}

		}()
		w.WriteHeader(http.StatusAccepted)
	}
}
