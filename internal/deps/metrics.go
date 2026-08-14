package deps

import "github.com/prometheus/client_golang/prometheus"

// Metrics is the metrics surface used by the internal packages. It matches the
// call shape produced by the pkg/metrics Series helpers, e.g.:
//
//	mc.Inc(series.Error("get_operation"))
//
// The production registry (metrics.Registry) satisfies this interface.
type Metrics interface {
	Inc(name string, labels prometheus.Labels)
	RecordDuration(name string, labels prometheus.Labels, duration float64)
}

// MetricsStub is a no-op Metrics for use in unit tests.
type MetricsStub struct{}

func NewMetricsStub() *MetricsStub { return &MetricsStub{} }

func (MetricsStub) Inc(string, prometheus.Labels)                     {}
func (MetricsStub) RecordDuration(string, prometheus.Labels, float64) {}
