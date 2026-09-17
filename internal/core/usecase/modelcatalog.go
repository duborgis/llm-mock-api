package usecase

import (
	"context"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
	"github.com/duborgis/llm-mock-api/internal/core/ports"
)

// ModelCatalog implements ports.ModelCatalogUseCase.
type ModelCatalog struct {
	Models ports.ModelRepository
}

func NewModelCatalog(models ports.ModelRepository) *ModelCatalog {
	return &ModelCatalog{Models: models}
}

func (m *ModelCatalog) List(ctx context.Context) ([]domain.ModelInfo, error) {
	return m.Models.List(ctx)
}

func (m *ModelCatalog) Get(ctx context.Context, id string) (domain.ModelInfo, error) {
	return m.Models.Get(ctx, id)
}
