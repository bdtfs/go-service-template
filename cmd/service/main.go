package main

import (
	"context"
	"log"

	"github.com/bdtfs/go-service-template/internal/config"
	"github.com/bdtfs/go-service-template/internal/di"
	"github.com/bdtfs/go-service-template/pkg/postgres"
	"github.com/bdtfs/go-service-template/pkg/service"
)

func main() {
	cfg := config.Must(config.Load("config.yaml"))

	var opts []service.Option

	// Compose infrastructure components based on config
	if cfg.Components.Postgres.Enabled {
		opts = append(opts, service.WithComponent(
			postgres.NewComponent(cfg.Components.Postgres.DSN),
		))
	}

	svc := service.Must(service.New(cfg, opts...))

	// Wire application-layer dependencies
	c := di.New(svc)
	_ = c // use c to register handlers, e.g.:
	// svc.HandleFunc("GET /api/v1/items", c.ItemHandler().List)

	if err := svc.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
