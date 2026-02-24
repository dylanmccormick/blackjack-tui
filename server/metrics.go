package server

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	ActiveTables      prometheus.Gauge
	ConnectedClients  prometheus.Gauge
	ConnectedDuration prometheus.Histogram

	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestsDuration *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		ActiveTables: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "blackjack_tables_active",
			Help: "Current number of active (living) tables",
		}),
		ConnectedClients: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "blackjack_connections_active",
			Help: "Current number of connected clients",
		}),
		ConnectedDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "blackjack_connection_duration_seconds",
			Help:    "Duration of connections for users grouped in buckets",
			Buckets: []float64{60, 300, 600, 1800, 3600, 7200},
		}),
		HTTPRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "blackjack_http_requests_total",
			Help: "Total http requests broken down by method, path, and statuscode",
		},
			[]string{"method", "path", "status"},
		),
		HTTPRequestsDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "blackjack_http_requests_duration_seconds",
			Help:    "Duration of requests broken down by method and path",
			Buckets: prometheus.DefBuckets,
		},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "blackjack_http_requests_in_flight",
			Help: "Number of currently executing http requests",
		}),
	}

	reg.MustRegister(
		m.ActiveTables,
		m.ConnectedClients,
		m.ConnectedDuration,
		m.HTTPRequestsTotal,
		m.HTTPRequestsDuration,
		m.HTTPRequestsInFlight,
	)
	return m
}
