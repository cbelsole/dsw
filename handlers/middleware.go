package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"
)

// RecoveryMiddleware recovers from panics thrown in your handlers by sending a
// 500 back with an error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}

				writeHTTPError(w, http.StatusInternalServerError, err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs the URI, Method, and time of your request
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("URI: %s, Method: %s, Time: %s\n", r.RequestURI, r.Method, time.Since(start))
	})
}
