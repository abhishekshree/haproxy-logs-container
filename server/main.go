package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var PORT string

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

func WithLogging(h http.Handler) http.Handler {
	loggingFn := func(rw http.ResponseWriter, req *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lrw := loggingResponseWriter{
			ResponseWriter: rw,
			responseData:   responseData,
		}
		h.ServeHTTP(&lrw, req)

		duration := time.Since(start)

		logrus.WithFields(logrus.Fields{
			"port":     PORT,
			"uri":      req.RequestURI,
			"method":   req.Method,
			"status":   responseData.status,
			"duration": duration,
			"size":     responseData.size,
		}).Info("request completed")
	}
	return http.HandlerFunc(loggingFn)
}

func indexHandler() http.Handler {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(rw, "Server is up.\n")
	}
	return http.HandlerFunc(fn)
}

func pingHandler() http.Handler {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(rw, "pong\n")
	}
	return http.HandlerFunc(fn)
}

func main() {
	PORT = os.Getenv("PORT")
	PORT = ":" + PORT

	f, err := os.OpenFile("./logs/webserver.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}
	logrus.SetOutput(f)

	http.Handle("/", WithLogging(indexHandler()))
	http.Handle("/ping", WithLogging(pingHandler()))

	logrus.WithField("PORT", PORT).Info("starting server")
	if err := http.ListenAndServe(PORT, nil); err != nil {
		logrus.WithField("event", "start server").Fatal(err)
	}
}
