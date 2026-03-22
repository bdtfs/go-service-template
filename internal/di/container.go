package di

import (
	"github.com/bdtfs/go-service-template/pkg/clog"
	"github.com/bdtfs/go-service-template/pkg/metrics"
	"github.com/bdtfs/go-service-template/pkg/postgres"
	"github.com/bdtfs/go-service-template/pkg/service"
	"github.com/bdtfs/go-service-template/pkg/transactions"
)

// Container provides lazy-initialized, application-layer dependency wiring.
// Infrastructure components (postgres, redis, etc.) live in the Service;
// the Container wires them into your domain's ports and use cases.
//
// Usage in main.go:
//
//	svc := service.Must(service.New(cfg, ...))
//	c := di.New(svc)
//	svc.HandleFunc("GET /api/v1/items", c.ItemHandler().List)
type Container struct {
	svc *service.Service

	// Add your application-layer singletons here, e.g.:
	// itemRepo    ports.ItemRepository
	// itemUseCase *usecase.ItemUseCase
	// itemHandler *handler.ItemHandler

	txFactory *transactions.PgTransactionFactory
	txManager transactions.TransactionManager
}

// New creates a Container backed by the given Service.
func New(svc *service.Service) *Container {
	return &Container{svc: svc}
}

// --- Core accessors (always available) ---

// Logger returns the service logger.
func (c *Container) Logger() clog.CLog {
	return c.svc.Logger()
}

// Metrics returns the metrics registry.
func (c *Container) Metrics() metrics.Registry {
	return c.svc.Metrics()
}

// --- Postgres helpers (available when postgres component is enabled) ---

// Postgres returns the postgres component, or nil if not registered.
func (c *Container) Postgres() *postgres.Component {
	comp, ok := c.svc.Component(postgres.ComponentName)
	if !ok {
		return nil
	}
	pg, _ := comp.(*postgres.Component)
	return pg
}

// TxFactory returns a lazy-initialized transaction factory.
func (c *Container) TxFactory() *transactions.PgTransactionFactory {
	return get(&c.txFactory, func() *transactions.PgTransactionFactory {
		pg := c.Postgres()
		if pg == nil || pg.Pool() == nil {
			return nil
		}
		f, _ := transactions.NewPgTransactionFactory(pg.Pool()).(*transactions.PgTransactionFactory)
		return f
	})
}

// TxManager returns a lazy-initialized transaction manager.
func (c *Container) TxManager() transactions.TransactionManager {
	return get(&c.txManager, func() transactions.TransactionManager {
		f := c.TxFactory()
		if f == nil {
			return nil
		}
		return transactions.NewPgTransactionManager(f)
	})
}

// --- Template: uncomment and adapt for your domain ---
//
// func (c *Container) ItemRepo() ports.ItemRepository {
// 	return get(&c.itemRepo, func() ports.ItemRepository {
// 		return adapters.NewItemPostgresRepo(c.Postgres().Pool(), c.TxFactory())
// 	})
// }
//
// func (c *Container) ItemUseCase() *usecase.ItemUseCase {
// 	return get(&c.itemUseCase, func() *usecase.ItemUseCase {
// 		return usecase.NewItemUseCase(c.ItemRepo(), c.TxManager(), c.Logger())
// 	})
// }
//
// func (c *Container) ItemHandler() *handler.ItemHandler {
// 	return get(&c.itemHandler, func() *handler.ItemHandler {
// 		return handler.NewItemHandler(c.ItemUseCase(), c.Logger(), c.Metrics())
// 	})
// }

// get is a generic lazy-initialization helper.
// On first call it runs builder and caches the result; subsequent calls return the cached value.
func get[T comparable](obj *T, builder func() T) T {
	if *obj != *new(T) {
		return *obj
	}
	*obj = builder()
	return *obj
}
