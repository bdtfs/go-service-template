package metrics

import (
	"net/http"
	"sync/atomic"
)

// HealthChecker manages liveness and readiness state for Kubernetes probes.
type HealthChecker struct {
	isReady   atomic.Bool
	isHealthy atomic.Bool
}

// NewHealthChecker creates a new HealthChecker (healthy=true, ready=false).
func NewHealthChecker() *HealthChecker {
	hc := &HealthChecker{}
	hc.isHealthy.Store(true)
	return hc
}

func (hc *HealthChecker) SetReady(ready bool) {
	hc.isReady.Store(ready)
}

func (hc *HealthChecker) SetHealthy(healthy bool) {
	hc.isHealthy.Store(healthy)
}

func (hc *HealthChecker) LivenessHandler(w http.ResponseWriter, _ *http.Request) {
	if hc.isHealthy.Load() {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte("unhealthy"))
}

func (hc *HealthChecker) ReadinessHandler(w http.ResponseWriter, _ *http.Request) {
	if hc.isReady.Load() {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	}

	w.WriteHeader(http.StatusServiceUnavailable)
	_, _ = w.Write([]byte("not ready"))
}
