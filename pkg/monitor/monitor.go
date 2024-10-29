package monitor

import "github.com/prometheus/client_golang/prometheus"

type MetricType int

const (
	Counter MetricType = iota
	Gauge
)

type Monitor interface {
	ID() string
	RegisterMetrics(registry *prometheus.Registry)
}

func NewMetric(registry *prometheus.Registry, typ MetricType, name, help string, executor func() float64) {
	switch typ {
	case Counter:
		registry.MustRegister(prometheus.NewCounterFunc(prometheus.CounterOpts{
			Name: name,
			Help: help,
		}, executor))
	case Gauge:
		registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		}, executor))
	}
}
