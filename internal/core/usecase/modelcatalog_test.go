package usecase

import (
	"context"
	"testing"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func TestModelCatalog_ListAndGet(t *testing.T) {
	repo := &fakeModelRepo{models: []domain.ModelInfo{{ID: "gpt-4o", Owner: "openai", Family: "gpt"}}}
	cat := NewModelCatalog(repo)

	list, err := cat.List(context.Background())
	if err != nil || len(list) != 1 {
		t.Fatalf("unexpected list result: %v %v", list, err)
	}

	m, err := cat.Get(context.Background(), "gpt-4o")
	if err != nil || m.Owner != "openai" {
		t.Fatalf("unexpected get result: %v %v", m, err)
	}

	_, err = cat.Get(context.Background(), "missing")
	if err != domain.ErrModelNotFound {
		t.Errorf("expected ErrModelNotFound, got %v", err)
	}
}
