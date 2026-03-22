package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bdtfs/go-service-template/internal/config"
	"github.com/bdtfs/go-service-template/pkg/clog"
	"github.com/bdtfs/go-service-template/pkg/metrics"
	"github.com/bdtfs/go-service-template/pkg/middleware"
)

// Type defines the service type which determines its runtime behavior.
type Type string

const (
	TypeAPI      Type = "api"
	TypeConsumer Type = "consumer"
	TypeWorker   Type = "worker"
)

// Service is the main application container that manages components,
// HTTP routing, metrics, and graceful lifecycle.
type Service struct {
	cfg        *config.Config
	logger     clog.CLog
	registry   metrics.Registry
	metricsSrv metrics.Server

	router     *http.ServeMux
	httpServer *http.Server
	middleware []func(http.Handler) http.Handler

	components     []Component
	componentIndex map[string]Component

	startFns []func(ctx context.Context) error

	serviceType Type
}

// Option configures a Service during construction.
type Option func(*Service) error

// New creates a new Service from the given config and options.
func New(cfg *config.Config, opts ...Option) (*Service, error) {
	logger := clog.NewCLog(cfg.Log.SlogLevel(), cfg.Log.Writer(), cfg.Log.AddSource)

	var registry metrics.Registry
	var metricsSrv metrics.Server

	if cfg.Metrics.Enabled {
		registry = metrics.NewRegistry(cfg.Metrics.Namespace, cfg.Metrics.Subsystem)
		metricsSrv = metrics.NewServer(cfg.Metrics.Address, logger, registry)
	} else {
		registry = metrics.NewRegistryStub()
		metricsSrv = metrics.NewServerStub()
	}

	svc := &Service{
		cfg:            cfg,
		logger:         logger,
		registry:       registry,
		metricsSrv:     metricsSrv,
		router:         http.NewServeMux(),
		componentIndex: make(map[string]Component),
		serviceType:    Type(cfg.Service.Type),
		middleware: []func(http.Handler) http.Handler{
			middleware.Recovery(logger),
			middleware.RequestID(),
			middleware.Logging(logger),
			middleware.Metrics(registry),
		},
	}

	for _, opt := range opts {
		if err := opt(svc); err != nil {
			return nil, err
		}
	}

	return svc, nil
}

// Must is like New but exits on error.
func Must(svc *Service, err error) *Service {
	if err != nil {
		slog.Error("failed to create service", "error", err)
		os.Exit(1)
	}
	return svc
}

// Logger returns the service logger.
func (s *Service) Logger() clog.CLog { return s.logger }

// Metrics returns the metrics registry.
func (s *Service) Metrics() metrics.Registry { return s.registry }

// Config returns the service configuration.
func (s *Service) Config() *config.Config { return s.cfg }

// Type returns the service type.
func (s *Service) Type() Type { return s.serviceType }

// Router returns the HTTP mux for direct route registration.
func (s *Service) Router() *http.ServeMux { return s.router }

// Component retrieves a registered component by name.
func (s *Service) Component(name string) (Component, bool) {
	c, ok := s.componentIndex[name]
	return c, ok
}

// Handle registers an HTTP route on the service router.
func (s *Service) Handle(pattern string, handler http.Handler) {
	s.router.Handle(pattern, handler)
}

// HandleFunc registers an HTTP handler function on the service router.
func (s *Service) HandleFunc(pattern string, handler http.HandlerFunc) {
	s.router.HandleFunc(pattern, handler)
}

// Run starts the service and blocks until a shutdown signal (SIGINT/SIGTERM) is received.
// It initializes all components, starts servers, and handles graceful shutdown.
func (s *Service) Run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			s.logger.ErrorCtx(ctx, fmt.Errorf("panic: %v", r), "recovered from panic")
		}
	}()

	// Start metrics server
	s.metricsSrv.Start(ctx)

	// Initialize components
	for _, c := range s.components {
		if err := c.Init(ctx); err != nil {
			return fmt.Errorf("initializing component %s: %w", c.Name(), err)
		}
		s.logger.InfoCtx(ctx, "component initialized: %s", c.Name())
	}

	// Mark service as ready
	s.metricsSrv.SetReady(true)
	s.metricsSrv.SetAlive(true)

	// Start HTTP server for API services
	if s.serviceType == TypeAPI {
		handler := s.buildHandler()

		s.httpServer = &http.Server{
			Addr:              s.cfg.Server.Port,
			Handler:           handler,
			ReadTimeout:       s.cfg.Server.ReadTimeout.Std(),
			WriteTimeout:      s.cfg.Server.WriteTimeout.Std(),
			IdleTimeout:       s.cfg.Server.IdleTimeout.Std(),
			ReadHeaderTimeout: time.Second,
		}

		go func() {
			s.logger.InfoCtx(ctx, "HTTP server listening on %s", s.cfg.Server.Port)
			if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.logger.ErrorCtx(ctx, err, "HTTP server error")
			}
		}()
	}

	// Run custom start functions (for consumers, workers, etc.)
	for _, fn := range s.startFns {
		go func() {
			if err := fn(ctx); err != nil {
				s.logger.ErrorCtx(ctx, err, "start function error")
			}
		}()
	}

	s.logger.InfoCtx(ctx, "service started: %s (%s)", s.cfg.Service.Name, s.serviceType)

	// Block until shutdown signal
	<-ctx.Done()
	s.logger.InfoCtx(context.Background(), "shutting down...")

	return s.shutdown()
}

// buildHandler wraps the router with the middleware stack.
func (s *Service) buildHandler() http.Handler {
	var handler http.Handler = s.router
	// Apply middleware in reverse so the first middleware added is the outermost.
	for i := len(s.middleware) - 1; i >= 0; i-- {
		handler = s.middleware[i](handler)
	}
	return handler
}

func (s *Service) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var errs []error

	// Mark as not ready
	s.metricsSrv.SetReady(false)

	// Shutdown HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("http server: %w", err))
		}
	}

	// Close components in reverse initialization order
	for i := len(s.components) - 1; i >= 0; i-- {
		if err := s.components[i].Close(ctx); err != nil {
			errs = append(errs, fmt.Errorf("closing %s: %w", s.components[i].Name(), err))
		}
		s.logger.InfoCtx(ctx, "component closed: %s", s.components[i].Name())
	}

	// Stop metrics server last
	if err := s.metricsSrv.Stop(ctx); err != nil {
		errs = append(errs, fmt.Errorf("metrics server: %w", err))
	}

	return errors.Join(errs...)
}

// --- Options ---

// WithComponent registers an infrastructure component with the service.
func WithComponent(c Component) Option {
	return func(s *Service) error {
		s.components = append(s.components, c)
		s.componentIndex[c.Name()] = c
		return nil
	}
}

// WithMiddleware appends HTTP middleware to the stack.
// Middleware is applied in order: first added = outermost wrapper.
func WithMiddleware(mw ...func(http.Handler) http.Handler) Option {
	return func(s *Service) error {
		s.middleware = append(s.middleware, mw...)
		return nil
	}
}

// WithStartFunc registers a function to run in a goroutine after service startup.
// Use this for consumer loops, background workers, etc.
func WithStartFunc(fn func(ctx context.Context) error) Option {
	return func(s *Service) error {
		s.startFns = append(s.startFns, fn)
		return nil
	}
}
