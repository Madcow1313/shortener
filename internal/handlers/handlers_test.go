package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
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
	for _, tt := range positiveTests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("google.com"))
			w := httptest.NewRecorder()
			f := HandleMainPage(&SimpleServer{Host: "213", BaseURL: "/", URLmap: map[string]string{}}, chi.NewRouter())
			f(w, request)
			res := w.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.want.code {
				t.Errorf("Error: wrong response - want %v, got %v in %v", tt.want.code, res.StatusCode, tt.name)
			}
			if res.ContentLength == 0 {
				t.Errorf("Error: no body in response")
			}
		})
	}

	for _, tt := range negativeTests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/", strings.NewReader("google.com"))
			w := httptest.NewRecorder()
			f := HandleMainPage(&SimpleServer{Host: "213", BaseURL: "/", URLmap: map[string]string{}}, chi.NewRouter())
			f(w, request)
			res := w.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.want.code {
				t.Errorf("Error: wrong response - want %v, got %v in %v", tt.want.code, res.StatusCode, tt.name)
			}
		})
	}
}

func TestHandleGetID(t *testing.T) {
	s := SimpleServer{
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
		s      SimpleServer
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
	for _, tt := range positiveTests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, "/"+tt.URL, strings.NewReader(tt.URL))
			w := httptest.NewRecorder()
			f := HandleGetID(&s, tt.URL, s.URLmap[tt.URL])
			f(w, request)
			res := w.Result()
			defer res.Body.Close()
			if res.StatusCode != tt.want.code {
				t.Errorf("Error: wrong response code - want %v, got %v in %v", tt.want.code, res.StatusCode, tt.name)
			}
			if loc := res.Header.Get("Location"); loc != tt.want.location {
				t.Errorf("Error: wrong URL- want %v, got %v in %v", tt.want.location, loc, tt.name)
			}
		})
	}
}
