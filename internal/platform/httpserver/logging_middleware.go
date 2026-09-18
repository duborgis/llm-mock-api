package httpserver

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder captures the status code written by the wrapped handler, since
// http.ResponseWriter doesn't expose it after the fact.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Flush forwards to the underlying ResponseWriter's http.Flusher, if it has one. Without this,
// wrapping a streaming handler (SSE chat/responses completions, chunked Vertex streaming) in
// this middleware breaks every `w.(http.Flusher)` type assertion those handlers rely on, since
// embedding the http.ResponseWriter interface does not promote methods (like Flush) that
// aren't part of that interface itself.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// WithAccessLog wraps handler with request/response logging: method, path, status, and
// latency for every request that hits this server. Without this, the mock ran completely
// silent — only startup/shutdown/fatal errors were ever logged.
func WithAccessLog(logger *slog.Logger, name string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		handler.ServeHTTP(rec, r)
		logger.Info("request",
			"server", name,
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"remote", r.RemoteAddr,
		)
	})
}
