package di

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/bdtfs/go-service-template/internal/deps"
	"github.com/bdtfs/go-service-template/internal/handler/create_operation"
	"github.com/bdtfs/go-service-template/internal/handler/get_operation"
	"github.com/bdtfs/go-service-template/internal/pkg/clients/notifier"
	"github.com/bdtfs/go-service-template/internal/pkg/storage"
	"github.com/bdtfs/go-service-template/internal/pkg/storage/postgres"
	"github.com/bdtfs/go-service-template/internal/pkg/usecase/operation"
	pgcomp "github.com/bdtfs/go-service-template/pkg/postgres"
	"github.com/bdtfs/go-service-template/pkg/service"
)

// Container lazily constructs the private application graph.
type Container struct {
	svc *service.Service

	trf         deps.TransactionFactory
	trm         deps.TransactionManager
	storage     *postgres.Storage
	notifier    *notifier.Client
	operationUC *operation.UseCase

	createOperationHandler *create_operation.Handler
	getOperationHandler    *get_operation.Handler
}

func New(svc *service.Service) *Container { return &Container{svc: svc} }

func (c *Container) logger() deps.Log { return c.svc.Logger() }

func (c *Container) metrics() deps.Metrics { return c.svc.Metrics() }

func (c *Container) pool() *pgxpool.Pool {
	comp, ok := c.svc.Component(pgcomp.ComponentName)
	if !ok {
		return nil
	}
	pg, ok := comp.(*pgcomp.Component)
	if !ok || pg == nil {
		return nil
	}
	return pg.Pool()
}

func (c *Container) TxFactory() deps.TransactionFactory {
	return get(&c.trf, func() deps.TransactionFactory {
		return storage.NewTransactionFactory(c.pool())
	})
}

func (c *Container) TxManager() deps.TransactionManager {
	return get(&c.trm, func() deps.TransactionManager {
		return storage.NewTransactionManager(c.pool())
	})
}

func (c *Container) Storage() *postgres.Storage {
	return get(&c.storage, func() *postgres.Storage {
		return postgres.NewStorage(c.TxFactory())
	})
}

func (c *Container) notifierClient() *notifier.Client {
	return get(&c.notifier, func() *notifier.Client {
		cfg := c.svc.Config().App.Notifier
		return notifier.New(c.metrics(), &http.Client{Timeout: cfg.Timeout.Std()}, cfg.BaseURL)
	})
}

func (c *Container) operationUseCase() *operation.UseCase {
	return get(&c.operationUC, func() *operation.UseCase {
		return operation.New(
			c.TxManager(),
			c.logger(),
			c.metrics(),
			c.Storage(),
			c.notifierClient(),
			uuid.New,
		)
	})
}

func (c *Container) CreateOperationHandler() *create_operation.Handler {
	return get(&c.createOperationHandler, func() *create_operation.Handler {
		return create_operation.New(c.logger(), c.metrics(), c.operationUseCase())
	})
}

func (c *Container) GetOperationHandler() *get_operation.Handler {
	return get(&c.getOperationHandler, func() *get_operation.Handler {
		return get_operation.New(c.logger(), c.metrics(), c.operationUseCase())
	})
}

func get[T comparable](obj *T, builder func() T) T {
	if *obj != *new(T) {
		return *obj
	}
	*obj = builder()
	return *obj
}
