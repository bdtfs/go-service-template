package main

import (
	"context"
	"log"

	"github.com/bdtfs/go-service-template/internal/config"
	"github.com/bdtfs/go-service-template/pkg/postgres"
	"github.com/bdtfs/go-service-template/pkg/service"
)

func main() {
	cfg := config.Must(config.Load("config.yaml"))

	var opts []service.Option

	// Compose components based on config
	if cfg.Components.Postgres.Enabled {
		opts = append(opts, service.WithComponent(
			postgres.NewComponent(cfg.Components.Postgres.DSN),
		))
	}

	svc := service.Must(service.New(cfg, opts...))

	// Register your routes here:
	// svc.HandleFunc("GET /api/v1/example", exampleHandler)

	if err := svc.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
