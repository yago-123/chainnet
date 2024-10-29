package monitor

import (
	"context"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"
	"net/http"
	"time"
)

const (
	MetricsPath                       = "/metrics"
	PrometheusExporterShutdownTimeout = 10 * time.Second
)

type PromExporter struct {
	logger   *logrus.Logger
	monitors []Monitor
	r        *httprouter.Router
	srv      *http.Server
	isActive bool
	registry *prometheus.Registry
	cfg      *config.Config
}

func NewPrometheusExporter(cfg *config.Config, monitors []Monitor) *PromExporter {
	r := httprouter.New()
	registry := prometheus.NewRegistry()

	r.GET(MetricsPath, func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			ErrorLog:      cfg.Logger,
			Timeout:       5 * time.Second,
			ErrorHandling: promhttp.ContinueOnError,
		}).ServeHTTP(w, req)
	})

	return &PromExporter{
		monitors: monitors,
		r:        r,
		registry: prometheus.NewRegistry(),
		logger:   cfg.Logger,
		cfg:      cfg,
	}
}

func (prom *PromExporter) Start() error {
	if prom.isActive {
		return nil
	}

	for _, monitor := range prom.monitors {
		monitor.RegisterMetrics(prom.registry)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", 8000),
		Handler:      prom.r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	prom.srv = srv
	prom.isActive = true

	go func() {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			prom.logger.Infof("prometheus exporter server stopped successfully")
		}

		if err != nil {
			prom.logger.Errorf("prometheus exporter server stopped with error: %s", err)
		}
	}()

	return nil
}

func (prom *PromExporter) Stop() error {
	if !prom.isActive {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), PrometheusExporterShutdownTimeout)
	defer cancel()

	if err := prom.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	prom.isActive = false
	return nil
}
