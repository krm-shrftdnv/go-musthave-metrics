package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func Initialize(level string) error {
	if Log != nil {
		return nil
	}
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl
	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	defer logger.Sync()

	Log = logger.Sugar()
	return nil
}

func RequestWithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		uri := r.RequestURI
		method := r.Method
		responseData := &responseData{
			status: http.StatusOK,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   responseData,
		}
		start := time.Now()
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)

		Log.Infow(
			"Request",
			"uri", uri,
			"method", method,
			"status", lw.responseData.status,
			"duration", duration,
			"size", lw.responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}
