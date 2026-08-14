package di_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bdtfs/go-service-template/internal/config"
	"github.com/bdtfs/go-service-template/internal/di"
	"github.com/bdtfs/go-service-template/pkg/service"
)

type registrar map[string]http.HandlerFunc

func (r registrar) HandleFunc(pattern string, handler http.HandlerFunc) {
	r[pattern] = handler
}

func TestRegisterRoutesDefersDependencyConstruction(t *testing.T) {
	cfg := &config.Config{}
	svc, err := service.New(cfg)
	require.NoError(t, err)
	c := di.New(svc)
	routes := registrar{}

	c.RegisterRoutes(routes)

	require.Len(t, routes, 2)
}
