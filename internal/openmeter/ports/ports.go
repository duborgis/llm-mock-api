package ports

import (
	"context"

	"github.com/duborgis/llm-mock-api/internal/openmeter/domain"
)

// EventIngestUseCase is the driving port: what the inbound HTTP adapter calls.
type EventIngestUseCase interface {
	Ingest(ctx context.Context, e domain.Event) error
	List(ctx context.Context, limit int) ([]domain.Event, error)
}

// EventRepository is the driven port: what the usecase needs from storage.
type EventRepository interface {
	Save(ctx context.Context, e domain.Event) error
	List(ctx context.Context, limit int) ([]domain.Event, error)
}
