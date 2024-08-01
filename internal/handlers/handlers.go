package handlers

import (
	"shortener/cmd/shortener/config"
	"shortener/internal/dbconnector"
	"shortener/internal/mylogger"
	server "shortener/internal/server/serverTypes"

	"github.com/go-chi/chi/v5"
)

type SimpleServer server.SimpleServer

const (
	userCookie           = "user_id"
	letters              = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	invalidRequestError  = "Invalid request method"
	parseFormError       = "Unable to parse form"
	bodyReadError        = "Unable to read body"
	selectShortError     = "Unable to get short url from database"
	unmarshalError       = "Unable to unmarshal json-data"
	marshalResponseError = "Unable to marshal response"
	pingError            = "Unable to connect to database"
)

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
	ZapLogger mylogger.Mylogger
	Router    *chi.Mux
	UserURLS  map[string][]string
}
