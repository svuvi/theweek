package middleware

import (
	"log"
	"net/http"
	"time"
)

type Logger struct {
	handler http.Handler
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	l.handler.ServeHTTP(rw, r)

	log.Printf("%d %s %s %v", rw.statusCode, r.Method, r.URL.Path, time.Since(start))
}

func NewLogger(handlerToWrap http.Handler) *Logger {
	return &Logger{handlerToWrap}
}
