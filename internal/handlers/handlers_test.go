package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	server "shortener/internal/server/serverTypes"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/assert/v2"
)

func TestHandleMainPage(t *testing.T) {
	type want struct {
		code int
	}
	positiveTests := []struct {
		name string
		want want
	}{
		{
			name: "positive test 1",
			want: want{
				code: 201,
			},
		},
		{
			name: "positive test 2",
			want: want{
				code: 201,
			},
		},
	}
	negativeTests := []struct {
		name       string
		method     string
		bodyReader *io.Reader
		want       want
	}{
		{
			name:   "negative test 1",
			method: http.MethodGet,
			want: want{
				code: 400,
			},
		},
		{
			name:       "negative test 2",
			method:     http.MethodPut,
			bodyReader: nil,
			want: want{
				code: 400,
			},
		},
	}
	var hh HandlerHelper
	hh.UserURLS = make(map[string][]string)
	for _, tt := range positiveTests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("google.com"))
			w := httptest.NewRecorder()
			hh.Server = &server.SimpleServer{Host: "213", BaseURL: "/", URLmap: map[string]string{},
				UserURLS: map[string][]string{}}
			hh.Router = chi.NewRouter()
			f := hh.HandlePostURL()
			f(w, request)
			res := w.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.want.code {
				t.Fatalf("Error: wrong response - want %v, got %v in %v", tt.want.code, res.StatusCode, tt.name)
			}
			if res.ContentLength == 0 {
				t.Fatalf("Error: no body in response")
			}
		})
	}

	for _, tt := range negativeTests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/", strings.NewReader("google.com"))
			w := httptest.NewRecorder()
			hh.Server = &server.SimpleServer{Host: "213", BaseURL: "/", URLmap: map[string]string{}}
			hh.Router = chi.NewRouter()
			f := hh.HandlePostURL()
			f(w, request)
			res := w.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.want.code {
				t.Fatalf("Error: wrong response - want %v, got %v in %v", tt.want.code, res.StatusCode, tt.name)
			}
		})
	}
}

func TestHandleGetID(t *testing.T) {
	s := server.SimpleServer{
		URLmap: map[string]string{
			"first":  "google.com",
			"second": "ya.ru",
			"third":  "http://example.com",
		},
		BaseURL: "/",
	}
	type want struct {
		code     int
		location string
	}
	positiveTests := []struct {
		name   string
		s      server.SimpleServer
		URL    string
		want   want
		method string
	}{
		{
			name:   "positive test 1",
			s:      s,
			URL:    "first",
			method: http.MethodGet,
			want: want{
				code:     307,
				location: "google.com",
			},
		},
		{
			name:   "positive test 2",
			s:      s,
			URL:    "second",
			method: http.MethodGet,
			want: want{
				code:     307,
				location: "ya.ru",
			},
		},
		{
			name:   "positive test 3",
			s:      s,
			URL:    "third",
			method: http.MethodGet,
			want: want{
				code:     307,
				location: "http://example.com",
			},
		},
	}
	var hh HandlerHelper
	for _, tt := range positiveTests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/"+tt.URL, strings.NewReader(tt.URL))
			w := httptest.NewRecorder()
			hh.Server = &s
			f := hh.HandleGetPostedURL(tt.URL, s.URLmap[tt.URL])
			f(w, request)
			res := w.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.want.code {
				t.Fatalf("Error: wrong response code - want %v, got %v in %v", tt.want.code, res.StatusCode, tt.name)
			}
			if !assert.IsEqual(tt.want.location, res.Header.Get("Location")) {
				t.Fatalf("Error: wrong URL- want %v, got %v in %v", tt.want.location, res.Header.Get("Location"), tt.name)
			}
		})
	}
}

func TestHandleApiShorten(t *testing.T) {
	type want struct {
		code     int
		response string
	}
	positiveTests := []struct {
		name string
		want want
	}{
		{
			name: "positive test 1",
			want: want{
				code: 201,
			},
		},
		{
			name: "positive test 2",
			want: want{
				code: 201,
			},
		},
	}
	var hh HandlerHelper
	hh.UserURLS = map[string][]string{}
	for _, tt := range positiveTests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`
			{
  				"url": "https://practicum.yandex.ru"
			} `))
			w := httptest.NewRecorder()
			hh.Server = &server.SimpleServer{Host: "213", BaseURL: "/", URLmap: map[string]string{}, UserURLS: map[string][]string{}}
			hh.Router = chi.NewRouter()
			f := hh.HandlePostAPIShorten()
			f(w, request)
			res := w.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.want.code {
				t.Fatalf("Error: wrong response - want %v, got %v in %v", tt.want.code, res.StatusCode, tt.name)
			}
			if res.ContentLength == 0 {
				t.Fatalf("Error: no body in response")
			}
		})
	}
}
