package main

import (
	"context"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

const (
	namespace = "http2pulsar" // For Prometheus metrics.
)

type Exporter struct {
	mutex        sync.Mutex
	receivedSns  prometheus.Counter
	sendSns      prometheus.Counter
	failedSns    prometheus.Counter
	sentDuration prometheus.Gauge
}

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

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.receivedSns.Describe(ch)
	e.sendSns.Describe(ch)
	e.failedSns.Describe(ch)
	e.sentDuration.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	e.receivedSns.Collect(ch)
	e.sendSns.Collect(ch)
	e.failedSns.Collect(ch)
	e.sentDuration.Collect(ch)
}

func sendMessage(e *Exporter) {
	begin := time.Now()
	time.Sleep(5 * time.Second)
	duration := time.Since(begin).Seconds()
	e.sendSns.Inc()
	e.sentDuration.Set(duration)
	// if err != nil {
	// 	e.failedSns.Inc()
	// 	return err
	// }
}

func main() {
	exporterName := namespace + "_exporter"
	exporter := NewExporter()
	prometheus.MustRegister(exporter)
	prometheus.MustRegister(version.NewCollector(exporterName))
	// Add Go module build info.
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		sendMessage(exporter)
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			log.Fatalf("testEndpoint: %s\n", err)
		}
	})

	disablePprof := false
	// Enable pprof endpoint if enabled by cmdline
	if !disablePprof {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Print("Server Started")

	log.Printf("SIGNAL %d received, then shutting down...\n", <-done)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		// extra handling here

		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")
}
