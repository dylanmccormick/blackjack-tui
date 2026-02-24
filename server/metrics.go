package server

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	ActiveTables      prometheus.Gauge
	ConnectedClients  prometheus.Gauge
	ConnectedDuration prometheus.Histogram
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
	}

	reg.MustRegister(m.ActiveTables)
	reg.MustRegister(m.ConnectedClients)
	reg.MustRegister(m.ConnectedDuration)
	return m
}
