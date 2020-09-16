package main

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "http2pulsar" // For Prometheus metrics.
)

// Exporter is prometheus exporter
type Exporter struct {
	mutex        sync.Mutex
	receivedSns  prometheus.Counter
	sendSns      prometheus.Counter
	failedSns    prometheus.Counter
	sentDuration prometheus.Gauge
}

// NewExporter create a Exporter
func NewExporter() *Exporter {
	return &Exporter{
		receivedSns: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "received_sns_total",
			Help:      "Total number of received sns.",
		}),
		sendSns: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "sent_sns_total",
			Help:      "Total number of processed sns sent to pulsar broker.",
		}),
		failedSns: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "failed_sns_total",
			Help:      "Total number of processed sns which failed on send to pulsar broker.",
		}),
		sentDuration: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "sent_sns_duration_seconds",
				Help: "Duration of sns send calls to the pulsar broker.",
			},
		),
	}
}

// Describe describe metrics
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.receivedSns.Describe(ch)
	e.sendSns.Describe(ch)
	e.failedSns.Describe(ch)
	e.sentDuration.Describe(ch)
}

// Collect collect metrics
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	e.receivedSns.Collect(ch)
	e.sendSns.Collect(ch)
	e.failedSns.Collect(ch)
	e.sentDuration.Collect(ch)
}
