package ports

import (
	"context"
	"time"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// ScenarioRepository is the driven port usecases use to resolve mock behavior for a given key.
// Implemented by adapters/outbound/scenario (in-memory store + YAML loader) and by admin HTTP writes.
type ScenarioRepository interface {
	Resolve(ctx context.Context, key domain.ScenarioKey) (domain.Scenario, error)
	Put(ctx context.Context, scenario domain.Scenario) error
	List(ctx context.Context) ([]domain.Scenario, error)
}

// ModelRepository is the driven port for the model catalog's backing data.
type ModelRepository interface {
	List(ctx context.Context) ([]domain.ModelInfo, error)
	Get(ctx context.Context, id string) (domain.ModelInfo, error)
}

// Clock is the driven port abstracting time, so latency/streaming timing is deterministic in tests.
type Clock interface {
	Now() time.Time
	Sleep(d time.Duration)
}
