package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// --- Registry Stub ---

type registryStub struct{}

func NewRegistryStub() Registry {
	return &registryStub{}
}

func (s *registryStub) Inc(_ string, _ prometheus.Labels) {}

func (s *registryStub) RecordDuration(_ string, _ prometheus.Labels, _ float64) {}

func (s *registryStub) PrometheusRegistry() *prometheus.Registry {
	return nil
}

// --- Server Stub ---

type serverStub struct{}

func NewServerStub() Server {
	return &serverStub{}
}

func (s *serverStub) Start(_ context.Context)          {}
func (s *serverStub) Stop(_ context.Context) error      { return nil }
func (s *serverStub) SetReady(_ bool)                   {}
func (s *serverStub) SetAlive(_ bool)                   {}
