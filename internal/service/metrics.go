package service

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// latency, traffic и errors
type httpMetrics struct {
	registry *prometheus.Registry
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func newHTTPMetrics(podID, datacenter string) *httpMetrics {
	labels := prometheus.Labels{
		"pod_id":     podID,
		"datacenter": datacenter,
	}
	m := &httpMetrics{
		registry: prometheus.NewRegistry(),
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        "http_requests_total",
			Help:        "Total HTTP requests.",
			ConstLabels: labels,
		}, []string{"handler", "method", "code"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:        "http_request_duration_seconds",
			Help:        "HTTP request duration in seconds.",
			ConstLabels: labels,
		}, []string{"handler", "method"}),
	}
	m.registry.MustRegister(m.requests, m.duration)
	return m
}

func (m *httpMetrics) register(mux *http.ServeMux, pattern string, handler http.HandlerFunc) {
	labels := prometheus.Labels{"handler": pattern}
	mux.Handle(pattern,
		promhttp.InstrumentHandlerDuration(m.duration.MustCurryWith(labels),
			promhttp.InstrumentHandlerCounter(m.requests.MustCurryWith(labels), handler)))
}

func (m *httpMetrics) scrapeHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
