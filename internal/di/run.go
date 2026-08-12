package di

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bdtfs/go-service-template/internal/config"
	pgcomp "github.com/bdtfs/go-service-template/pkg/postgres"
	"github.com/bdtfs/go-service-template/pkg/service"
)

func Run(ctx context.Context, configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	var opts []service.Option
	if cfg.Components.Postgres.Enabled {
		opts = append(opts, service.WithComponent(pgcomp.NewComponent(cfg.Components.Postgres.DSN)))
	}

	svc, err := service.New(cfg, opts...)
	if err != nil {
		return fmt.Errorf("build service: %w", err)
	}

	New(svc).RegisterRoutes(svc)
	return svc.Run(ctx)
}

type routeRegistrar interface {
	HandleFunc(string, http.HandlerFunc)
}

func (c *Container) RegisterRoutes(r routeRegistrar) {
	r.HandleFunc("POST /api/v1/operations", func(w http.ResponseWriter, req *http.Request) {
		c.CreateOperationHandler().Handle(w, req)
	})
	r.HandleFunc("GET /api/v1/operations/{external_id}", func(w http.ResponseWriter, req *http.Request) {
		c.GetOperationHandler().Handle(w, req)
	})
}
