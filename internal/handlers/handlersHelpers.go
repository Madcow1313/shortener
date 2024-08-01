package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"math/big"
	"net/http"
	"os"
	"shortener/cmd/shortener/config"
	server "shortener/internal/server/serverTypes"
	"strings"
)

func (hh *HandlerHelper) GetUserIDFromCookie(w http.ResponseWriter, r *http.Request) string {
	cookie := w.Header().Get("Set-Cookie")
	if cookie == "" {
		cookies, err := r.Cookie(userCookie)
		if err != nil {
			return ""
		}
		cookie = cookies.Value
	}
	return strings.TrimPrefix(cookie, userCookie+"=")
}

func (hh *HandlerHelper) AddUserURL(userID string, url string) {
	if _, ok := hh.UserURLS[userID]; !ok {
		hh.UserURLS[userID] = make([]string, 0)
	}
	hh.UserURLS[userID] = append(hh.UserURLS[userID], url)
}

func (hh *HandlerHelper) ShortenURL() string {
	letters := []rune(letters)
	result := make([]rune, 6)
	for i := 0; i < 6; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			hh.ZapLogger.LogError(err)
			continue
		}
		result[i] = letters[index.Int64()]
	}
	return string(result)
}

func (hh *HandlerHelper) WriteToStorage(shortened string, origin string, userID string) error {
	var err error
	switch hh.Config.StorageType {
	case config.Database:
		err = hh.Connector.ConnectToDB(func(db *sql.DB, args ...interface{}) error {
			return hh.Connector.InsertURL(db, shortened, origin, userID)
		})
	case config.File:
		err = WriteToFileStorage(hh.Server.Storage, origin, shortened, hh.Server.ID)

	default:
	}
	return err
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
