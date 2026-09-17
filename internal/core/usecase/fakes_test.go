package usecase

import (
	"context"
	"time"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// fakeScenarioRepo is an in-memory fake of ports.ScenarioRepository for unit tests.
type fakeScenarioRepo struct {
	scenarios map[string]domain.Scenario
}

func newFakeScenarioRepo() *fakeScenarioRepo {
	return &fakeScenarioRepo{scenarios: map[string]domain.Scenario{}}
}

func (f *fakeScenarioRepo) key(k domain.ScenarioKey) string { return k.Model + "|" + k.Operation }

func (f *fakeScenarioRepo) Resolve(ctx context.Context, k domain.ScenarioKey) (domain.Scenario, error) {
	if s, ok := f.scenarios[f.key(k)]; ok {
		return s, nil
	}
	return domain.Scenario{Key: k.Model}, nil
}

func (f *fakeScenarioRepo) Put(ctx context.Context, s domain.Scenario) error {
	f.scenarios[s.Key] = s
	return nil
}

func (f *fakeScenarioRepo) List(ctx context.Context) ([]domain.Scenario, error) {
	out := make([]domain.Scenario, 0, len(f.scenarios))
	for _, s := range f.scenarios {
		out = append(out, s)
	}
	return out, nil
}

// fakeClock is a deterministic, non-sleeping fake of ports.Clock for unit tests.
type fakeClock struct {
	now time.Time
}

func newFakeClock() *fakeClock { return &fakeClock{now: time.Unix(1700000000, 0)} }

func (c *fakeClock) Now() time.Time        { return c.now }
func (c *fakeClock) Sleep(d time.Duration) { c.now = c.now.Add(d) }

// fakeModelRepo is an in-memory fake of ports.ModelRepository for unit tests.
type fakeModelRepo struct {
	models []domain.ModelInfo
}

func (f *fakeModelRepo) List(ctx context.Context) ([]domain.ModelInfo, error) {
	return f.models, nil
}

func (f *fakeModelRepo) Get(ctx context.Context, id string) (domain.ModelInfo, error) {
	for _, m := range f.models {
		if m.ID == id {
			return m, nil
		}
	}
	return domain.ModelInfo{}, domain.ErrModelNotFound
}
