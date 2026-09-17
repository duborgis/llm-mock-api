package scenario

import (
	"context"
	"sync"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// Store implements ports.ScenarioRepository as an in-memory map, mutable at runtime
// via the admin HTTP API and seedable at startup from the YAML loader.
type Store struct {
	mu        sync.RWMutex
	scenarios map[string]domain.Scenario
}

func NewStore() *Store {
	return &Store{scenarios: map[string]domain.Scenario{}}
}

func key(k domain.ScenarioKey) string {
	op := k.Operation
	if op == "" {
		op = "chat"
	}
	return k.Model + "|" + op
}

func (s *Store) Resolve(ctx context.Context, k domain.ScenarioKey) (domain.Scenario, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if sc, ok := s.scenarios[key(k)]; ok {
		return sc, nil
	}
	// No scenario configured: return a benign default (deterministic canned reply, no error/latency).
	return domain.Scenario{Key: key(k)}, nil
}

// Put stores/replaces a scenario. Scenario.Key must already be in "model|operation" form
// (as produced by key()); callers going through the admin API build it that way.
func (s *Store) Put(ctx context.Context, sc domain.Scenario) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scenarios[sc.Key] = sc
	return nil
}

func (s *Store) List(ctx context.Context) ([]domain.Scenario, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Scenario, 0, len(s.scenarios))
	for _, sc := range s.scenarios {
		out = append(out, sc)
	}
	return out, nil
}

// Key exposes the store's key-building convention so other adapters (loader, admin) stay consistent.
func Key(model, operation string) string {
	return key(domain.ScenarioKey{Model: model, Operation: operation})
}
