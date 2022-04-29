package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricScrapeSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webscraper_scrape_success",
		Help: "Successful scrape",
	})
	metricScrapeError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webscraper_scrape_error",
		Help: "Scrape ended with error",
	})
)
