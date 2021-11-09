package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// AccessLog is a simple access-log middleware that emmits JSON messages to stdout.
func AccessLog(next http.Handler) http.Handler {
	type msg struct {
		RemoteAddr   string `json:"remote_addr"`
		Method       string `json:"method"`
		RequestURI   string `json:"request_uri"`
		StatusCode   int    `json:"status_code"`
		Duration     int64  `json:"duration_ms"`
		RequestSize  int64  `json:"request_size"`
		ResponseSize int    `json:"response_size"`
	}

	encoder := json.NewEncoder(os.Stdout)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wr := &wrapper{w, http.StatusOK, 0}
		next.ServeHTTP(wr, r)
		msg := msg{
			RemoteAddr:   r.RemoteAddr,
			Method:       r.Method,
			RequestURI:   r.RequestURI,
			StatusCode:   wr.statusCode,
			Duration:     time.Since(start).Milliseconds(),
			RequestSize:  r.ContentLength,
			ResponseSize: wr.size,
		}
		if shouldIgnoreLog(r.RequestURI) {
			// Don't care about those
			return
		}
		if err := encoder.Encode(&msg); err != nil {
			log.Printf("ERROR encoding access log: %v", err)
		}
	})
}

// wrapper to record status code and response size
type wrapper struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (w *wrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
func (w *wrapper) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func shouldIgnoreLog(uri string) bool {
	for _, u := range []string{
		"/favicon.png",
		"/bootstrap.min.css",
		"/bootstrap.min.css.map",
	} {
		if u == uri {
			return true
		}
	}
	return false
}
