package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"shortener/internal/mylogger"
	"strings"
)

type GzipResponseWriter struct {
	http.ResponseWriter
	ResponseData *mylogger.ResponseData
}

type GzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

type Compressor struct{}

func (grw *GzipResponseWriter) Write(b []byte) (int, error) {
	comp := gzip.NewWriter(grw.ResponseWriter)
	size, err := comp.Write(b)
	comp.Close()
	grw.ResponseData.Size += size
	return size, err
}

func (grw *GzipResponseWriter) WriteHeader(statusCode int) {
	grw.ResponseWriter.WriteHeader(statusCode)
	if statusCode < 300 {
		grw.Header().Set("Content-Encoding", "gzip")
	}
	grw.ResponseData.Status = statusCode
}

func (gr *GzipReader) NewGzipReader(r io.ReadCloser) (*GzipReader, error) {
	reader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &GzipReader{
		r:  r,
		zr: reader,
	}, nil
}

func (gr *GzipReader) Read(p []byte) (n int, err error) {
	return gr.zr.Read(p)
}

func (gr *GzipReader) Close() (err error) {
	if err = gr.r.Close(); err != nil {
		return err
	}
	return gr.zr.Close()
}

func (c *Compressor) Decompress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			h(w, r)
			return
		}
		var gzr GzipReader
		reader, err := gzr.NewGzipReader(r.Body)
		if err != nil {
			h(w, r)
			return
		}
		r.Body = reader.zr
		defer r.Body.Close()
		h(w, r)
	}
}

func (c *Compressor) Compress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h(w, r)
			return
		}
		grw := GzipResponseWriter{
			ResponseWriter: w,
			ResponseData:   &mylogger.ResponseData{},
		}
		h(&grw, r)
	}
}
