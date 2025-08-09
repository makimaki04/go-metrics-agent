package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func WithLogging(h http.HandlerFunc, logger *zap.Logger) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		uri := r.RequestURI
		method := r.Method
		headers := r.Header.Get("Accept-Encoding")

		responseData := &responseData {
			status: 0,
			size: 0,
		}

		lw := loggingResponseWriter {
			ResponseWriter: w,
			responseData: responseData,
		}

		h(&lw, r)
		
		duration := time.Since(start)

		logger.Sugar().Infoln(
			"uri", uri,
			"method", method,
			"headers", headers,
			"status", responseData.status,
			"size", responseData.size,
			"duration", duration,
		)
	}

	return logFn
}

type (
	responseData struct {
		status int
		size int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (l *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := l.ResponseWriter.Write(b)
	l.responseData.size = size
	return size, err
}

func (l *loggingResponseWriter) WriteHeader(statuceCode int) {
	l.ResponseWriter.WriteHeader(statuceCode)
	l.responseData.status = statuceCode
}