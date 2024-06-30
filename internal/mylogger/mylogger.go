package mylogger

import (
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type (
	// берём структуру для хранения сведений об ответе
	ResponseData struct {
		Status int
		Size   int
	}

	// добавляем реализацию http.ResponseWriter
	LoggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		ResponseData        *ResponseData
	}
)

var ZapLogger *zap.Logger = zap.L()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	ZapLogger = zl
	return nil
}

func LogRequest(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &ResponseData{
			Status: 0,
			Size:   0,
		}
		lw := LoggingResponseWriter{
			ResponseWriter: w,
			ResponseData:   responseData,
		}
		h(&lw, r)
		duration := time.Since(start)
		ZapLogger.Log(ZapLogger.Level(), "Incoming request",
			zap.String("method", r.Method),
			zap.String("URI", r.RequestURI),
			zap.String("duration", duration.String()),
		)
		status := strconv.FormatInt(int64(lw.ResponseData.Status), 10)
		size := strconv.FormatInt(int64(lw.ResponseData.Size), 10)
		ZapLogger.Log(ZapLogger.Level(), "Response",
			zap.String("status", status),
			zap.String("size", size),
		)
	})
}

func (r *LoggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.ResponseData.Size += size // захватываем размер
	return size, err
}

func (r *LoggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.ResponseData.Status = statusCode // захватываем код статуса
}
