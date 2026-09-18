package usecase

import (
	"context"
	"time"

	"github.com/duborgis/llm-mock-api/internal/openmeter/domain"
	"github.com/duborgis/llm-mock-api/internal/openmeter/ports"
)

// Ingest implements ports.EventIngestUseCase. Pure business logic: no HTTP, no MongoDB driver.
type Ingest struct {
	Events ports.EventRepository
}

func NewIngest(events ports.EventRepository) *Ingest {
	return &Ingest{Events: events}
}

func (u *Ingest) Ingest(ctx context.Context, e domain.Event) error {
	e.ReceivedAt = time.Now().UTC()
	return u.Events.Save(ctx, e)
}

func (u *Ingest) List(ctx context.Context, limit int) ([]domain.Event, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	return u.Events.List(ctx, limit)
}
