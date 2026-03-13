package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"github.com/Dhushyanthc/event-feed-engine/internal/contextutils"

	
)

type loggingResponseWriter struct {
	http.ResponseWriter 
	statusCode int
}


func (lrw *loggingResponseWriter) WriteHeader (code int) {
	lrw.statusCode = code 
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware (logger *zap.Logger, next http.HandlerFunc) http.HandlerFunc{
	return func (w http.ResponseWriter, r *http.Request){

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}


		start := time.Now()
		next(lrw,r)
		duration := time.Since(start)
		method := r.Method
		path := r.URL.Path
		status := lrw.statusCode
		requestID := contextutils.GetRequestID(r)
		
		logger.Info(
			"request completed",
			zap.String("request-id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", duration),
		)
	}
}