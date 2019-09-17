package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func serveMetrics(config Config) {
	address := config.MetricsAddress
	if address == "" {
		address = ":8080"
	}

	log.Printf("Serving prometheus metrics on %s", address)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(address, nil))
}
