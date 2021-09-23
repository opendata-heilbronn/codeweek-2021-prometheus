package main

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"strconv"
)

var apiHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "http_request_duration_seconds",
	Help: "Latency for handling requests",
}, []string{"handler", "method", "code"})

func init() {
	prometheus.MustRegister(apiHistogram)
}

func statusHandler(writer http.ResponseWriter, request *http.Request) {
	wantedStatus := request.URL.Query().Get("code")
	statusCode, err := strconv.ParseInt(wantedStatus, 10, 64)
	if err != nil {
		statusCode = 200
	}

	writer.WriteHeader(int(statusCode))
}

func randomHandler(writer http.ResponseWriter, request *http.Request) {
	sizeString := request.URL.Query().Get("bytes")
	size, err := strconv.ParseInt(sizeString, 10, 64)
	if err != nil {
		size = 100
	}

	buffer := make([]byte, size)
	_, err = rand.Reader.Read(buffer)
	if err != nil {
		http.Error(writer, err.Error(), 500)
		return
	}
	writer.WriteHeader(200)
	_, _ = writer.Write([]byte(base64.StdEncoding.EncodeToString(buffer)))
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/status", http.HandlerFunc(statusHandler))
	mux.Handle("/random", http.HandlerFunc(randomHandler))
	/*
		mux.Handle("/status", promhttp.InstrumentHandlerDuration(
			apiHistogram.MustCurryWith(prometheus.Labels{"handler": "status"}),
			http.HandlerFunc(statusHandler)),
		)
		mux.Handle("/random", promhttp.InstrumentHandlerDuration(
			apiHistogram.MustCurryWith(prometheus.Labels{"handler": "random"}),
			http.HandlerFunc(randomHandler)),
		)
	*/

	mux.Handle("/metrics", promhttp.Handler())

	srv := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Println("Listening on :8080")
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
