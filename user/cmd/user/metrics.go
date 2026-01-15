package main

import (
	"database/sql"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	MessagesProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_message_handler_processed_total",
			Help: "Total number of AMQP messages processed",
		},
		[]string{"status"},
	)

	MessageDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "user_message_handler_duration_seconds",
			Help:    "Time spent processing AMQP messages",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func registerMetrics(router *mux.Router, db *sql.DB) {
	registry := prometheus.NewRegistry()

	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	registry.MustRegister(MessagesProcessed)
	registry.MustRegister(MessageDuration)

	if db != nil {
		registry.MustRegister(collectors.NewDBStatsCollector(db, "user_db"))
	}

	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
}
