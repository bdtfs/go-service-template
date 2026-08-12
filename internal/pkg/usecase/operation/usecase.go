package operation

import (
	"context"

	"github.com/google/uuid"

	"github.com/bdtfs/go-service-template/internal/deps"
	"github.com/bdtfs/go-service-template/internal/model"
	"github.com/bdtfs/go-service-template/internal/pkg/clients/notifier"
	"github.com/bdtfs/go-service-template/pkg/metrics"
)

// storage is the persistence surface this use case needs. Interfaces are
// declared at the point of use so the use case owns its contract; mocks are
// generated with mockery (see internal/deps/generate.go).
//
//go:generate ../../../../bin/mockery --name storage
type storage interface {
	SaveOperation(ctx context.Context, op *model.Operation) error
	GetOperation(ctx context.Context, externalID string) (*model.Operation, error)
	UpdateOperationStatus(ctx context.Context, externalID string, status model.OperationStatus) error
}

//go:generate ../../../../bin/mockery --name notifierClient
type notifierClient interface {
	NotifyOperationCreated(ctx context.Context, in notifier.OperationCreatedIn) error
}

// UseCase orchestrates the operation domain: it validates inputs against
// business rules, persists state transactionally, and emits side effects.
type UseCase struct {
	trm      deps.TransactionManager
	log      deps.Log
	mc       deps.Metrics
	s        metrics.Series
	st       storage
	notifier notifierClient
	uuid     func() uuid.UUID
}

func New(
	trm deps.TransactionManager,
	log deps.Log,
	mc deps.Metrics,
	st storage,
	nc notifierClient,
	newUUID func() uuid.UUID,
) *UseCase {
	return &UseCase{
		trm:      trm,
		log:      log,
		mc:       mc,
		s:        metrics.NewSeries(metrics.SeriesTypeUseCase, "operation"),
		st:       st,
		notifier: nc,
		uuid:     newUUID,
	}
}
