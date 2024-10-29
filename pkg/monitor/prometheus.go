package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/yago-123/chainnet/config"
	"net/http"
)

type PromExporter struct {
	cfg      *config.Config
	monitors []Monitor
	registry *prometheus.Registry
}

func NewPrometheusExporter(cfg *config.Config, monitors []Monitor) *PromExporter {
	return &PromExporter{
		cfg:      cfg,
		monitors: monitors,
		registry: prometheus.NewRegistry(),
	}
}

func (prom *PromExporter) Start() {
	for _, monitor := range prom.monitors {
		monitor.RegisterMetrics(prom.registry)

	}

	http.Handle("/metrics", promhttp.HandlerFor(prom.registry, promhttp.HandlerOpts{}))

	// todo(): use config to set the port
	// todo(): move this to a modular, non-blocking initialization
	http.ListenAndServe(":8000", nil)
}

func (prom *PromExporter) Stop() {

}
