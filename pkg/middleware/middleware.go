package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/bdtfs/go-service-template/pkg/clog"
	"github.com/bdtfs/go-service-template/pkg/metrics"
)

type requestIDKey struct{}

// RequestIDFromContext extracts the request ID from the context.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// Recovery returns middleware that recovers from panics and logs the error.
func Recovery(logger clog.CLog) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.ErrorCtx(r.Context(), fmt.Errorf("panic: %v", rec), "recovered from panic in HTTP handler")
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID returns middleware that injects a unique request ID into the context
// and sets it as the X-Request-ID response header.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get("X-Request-ID")
			if id == "" {
				id = generateID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey{}, id)
			w.Header().Set("X-Request-ID", id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Logging returns middleware that logs each HTTP request with method, path,
// status code, and duration.
func Logging(logger clog.CLog) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			ctx := r.Context()

			if reqID := RequestIDFromContext(ctx); reqID != "" {
				ctx = logger.AddKeysValuesToCtx(ctx, map[string]any{
					"request_id": reqID,
				})
			}

			logger.InfoCtx(ctx, "%s %s %d %s",
				r.Method, r.URL.Path, rw.status, duration)
		})
	}
}

// Metrics returns middleware that records HTTP request count and duration
// using the provided metrics registry.
func Metrics(registry metrics.Registry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			labels := prometheus.Labels{
				"method": r.Method,
				"status": strconv.Itoa(rw.status),
			}
			registry.Inc("http_requests_total", labels)
			registry.RecordDuration("http_request_duration_seconds", labels, duration.Seconds())
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.wroteHeader = true
	}
	return rw.ResponseWriter.Write(b)
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
