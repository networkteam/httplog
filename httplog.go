package httplog

import (
	"net/http"
	"time"

	"github.com/apex/log"
)

// New middleware wrapping `h`.
func New(h http.Handler) *Logger {
	return &Logger{Handler: h}
}

// Logger middleware wrapping Handler.
type Logger struct {
	Handler http.Handler
}

// wrapper to capture status.
type wrapper struct {
	http.ResponseWriter
	written int
	status  int
}

// capture status.
func (w *wrapper) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// capture written bytes.
func (w *wrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.written += n
	return n, err
}

// ServeHTTP implementation.
func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	res := &wrapper{w, 0, 200}

	ctx := log.WithFields(log.Fields{
		"method": r.Method,
		"url":    r.RequestURI,
	})

	ctx.Info("request")
	l.Handler.ServeHTTP(res, r)

	ctx = ctx.WithFields(log.Fields{
		"status":   res.status,
		"size":     res.written,
		"duration": time.Since(start),
	})

	switch {
	case res.status >= 500:
		ctx.Error("response")
	case res.status >= 400:
		ctx.Warn("response")
	default:
		ctx.Info("response")
	}
}
