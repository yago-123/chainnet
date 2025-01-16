package monitor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/yago-123/chainnet/config"
)

const (
	ReadWriteTimeout = 5 * time.Second
	IdleTimeout      = 10 * time.Second

	PrometheusExporterShutdownTimeout = 10 * time.Second
)

type PromExporter struct {
	monitors []Monitor
	r        *httprouter.Router
	srv      *http.Server

	isActive bool
	registry *prometheus.Registry

	logger *logrus.Logger
	cfg    *config.Config
}

func NewPrometheusExporter(cfg *config.Config, monitors []Monitor) *PromExporter {
	r := httprouter.New()
	registry := prometheus.NewRegistry()

	r.GET(cfg.Prometheus.Path, func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			ErrorLog:      cfg.Logger,
			Timeout:       ReadWriteTimeout,
			ErrorHandling: promhttp.ContinueOnError,
		}).ServeHTTP(w, req)
	})

	// register the metrics for each monitor
	for _, monitor := range monitors {
		monitor.RegisterMetrics(registry)
	}

	return &PromExporter{
		monitors: monitors,
		r:        r,
		registry: registry,
		logger:   cfg.Logger,
		cfg:      cfg,
	}
}

func (prom *PromExporter) Start() error {
	if prom.isActive {
		return nil
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf("127.0.0.1:%d", prom.cfg.Prometheus.Port),
		Handler:      prom.r,
		ReadTimeout:  ReadWriteTimeout,
		WriteTimeout: ReadWriteTimeout,
		IdleTimeout:  IdleTimeout,
	}

	prom.srv = srv
	prom.isActive = true

	go func() {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			prom.logger.Infof("prometheus exporter server stopped successfully")
			return
		}

		if err != nil {
			prom.logger.Errorf("prometheus exporter server stopped with error: %s", err)
			return
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
