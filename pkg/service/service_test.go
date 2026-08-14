package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bdtfs/go-service-template/internal/config"
	"github.com/bdtfs/go-service-template/pkg/service"
)

func TestRunStopsWhenContextIsCancelled(t *testing.T) {
	cfg := &config.Config{}
	cfg.Service.Name = "test"
	cfg.Service.Type = string(service.TypeWorker)
	svc, err := service.New(cfg)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	require.NoError(t, svc.Run(ctx))
}
