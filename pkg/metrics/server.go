package metrics

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bdtfs/go-service-template/pkg/clog"
)

const (
	metricsEndpoint   = "/metrics"
	livenessEndpoint  = "/healthz"
	readinessEndpoint = "/readyz"
)

var (
	memAllocGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_mem_stats_alloc_bytes",
		Help: "Number of bytes allocated and still in use.",
	})
	memSysGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_mem_stats_sys_bytes",
		Help: "Number of bytes obtained from the system.",
	})
	memHeapAllocGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_mem_stats_heap_alloc_bytes",
		Help: "Number of heap bytes allocated and still in use.",
	})
	memHeapSysGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_mem_stats_heap_sys_bytes",
		Help: "Number of heap bytes obtained from the system.",
	})
)

type server struct {
	addr        string
	logger      clog.CLog
	registry    Registry
	healthCheck *HealthChecker
	httpServer  *http.Server
	stopCh      chan struct{}
}

// NewServer creates a metrics server that exposes /metrics, /healthz, and /readyz.
func NewServer(addr string, logger clog.CLog, registry Registry) Server {
	registry.PrometheusRegistry().MustRegister(memAllocGauge, memSysGauge, memHeapAllocGauge, memHeapSysGauge)

	return &server{
		addr:        addr,
		logger:      logger,
		registry:    registry,
		healthCheck: NewHealthChecker(),
		stopCh:      make(chan struct{}),
	}
}

func (s *server) Start(ctx context.Context) {
	mux := http.NewServeMux()

	mux.Handle(metricsEndpoint, promhttp.HandlerFor(s.registry.PrometheusRegistry(), promhttp.HandlerOpts{}))
	mux.HandleFunc(livenessEndpoint, s.healthCheck.LivenessHandler)
	mux.HandleFunc(readinessEndpoint, s.healthCheck.ReadinessHandler)

	s.httpServer = &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go s.collectMemoryStats(ctx)

	go func() {
		s.logger.InfoCtx(ctx, "metrics server started on %s", s.addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.ErrorCtx(ctx, err, "failed to start metrics server")
		}
	}()
}

func (s *server) Stop(ctx context.Context) error {
	close(s.stopCh)

	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}

func (s *server) SetReady(ready bool) {
	s.healthCheck.SetReady(ready)
}

func (s *server) SetAlive(alive bool) {
	s.healthCheck.SetHealthy(alive)
}

func (s *server) collectMemoryStats(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			memAllocGauge.Set(float64(memStats.Alloc))
			memSysGauge.Set(float64(memStats.Sys))
			memHeapAllocGauge.Set(float64(memStats.HeapAlloc))
			memHeapSysGauge.Set(float64(memStats.HeapSys))
		}
	}
}
