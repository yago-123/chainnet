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

// NewMetric registers a new metric with the given name and help string
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

// NewMetricWithLabelsAsync registers a new metric with an executor function that must be run asynchronously. This
// is required for modules that contain metrics that are hard to calculate synchronously. Only support for Gauge
func NewMetricWithLabelsAsync(
	registry *prometheus.Registry,
	typ MetricType,
	name, help string,
	labels []string,
	executor func(metricVec interface{}),
) {
	switch typ {
	case Counter:
		metric := prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: help,
		}, labels)
		registry.MustRegister(metric)
		go executor(metric)

	case Gauge:
		metric := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		}, labels)
		registry.MustRegister(metric)
		go executor(metric)
	}
}
