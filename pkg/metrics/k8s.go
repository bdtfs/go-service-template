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
		if _, err := w.Write([]byte("ok")); err != nil {
			return
		}
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	if _, err := w.Write([]byte("unhealthy")); err != nil {
		return
	}
}

func (hc *HealthChecker) ReadinessHandler(w http.ResponseWriter, _ *http.Request) {
	if hc.isReady.Load() {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			return
		}
		return
	}

	w.WriteHeader(http.StatusServiceUnavailable)
	if _, err := w.Write([]byte("not ready")); err != nil {
		return
	}
}
