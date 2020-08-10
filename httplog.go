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
	http.Handler
}

// ServeHTTP implementation.
func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	res := makeLoggingResponseWriter(w)

	ctx := log.WithFields(log.Fields{
		"url":        r.RequestURI,
		"method":     r.Method,
		"remoteAddr": r.RemoteAddr,
	})

	ctx.Info("request")
	l.Handler.ServeHTTP(res, r)

	ctx = ctx.WithFields(log.Fields{
		"status":   res.Status(),
		"size":     res.Size(),
		"duration": ms(time.Since(start)),
	})

	switch {
	case res.Status() >= 500:
		ctx.Error("response")
	case res.Status() >= 400:
		ctx.Warn("response")
	default:
		ctx.Info("response")
	}
}

// ms returns the duration in milliseconds.
func ms(d time.Duration) int {
	return int(d / time.Millisecond)
}
