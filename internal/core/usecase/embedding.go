package usecase

import (
	"context"
	"hash/fnv"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
	"github.com/duborgis/llm-mock-api/internal/core/ports"
)

// Embedding implements ports.EmbeddingUseCase.
type Embedding struct {
	Scenarios ports.ScenarioRepository
	Clock     ports.Clock
	Dims      int
}

func NewEmbedding(scenarios ports.ScenarioRepository, clock ports.Clock) *Embedding {
	return &Embedding{Scenarios: scenarios, Clock: clock, Dims: 8}
}

func (e *Embedding) Embed(ctx context.Context, req domain.EmbeddingRequest) (domain.EmbeddingResponse, error) {
	if req.Model == "" || len(req.Input) == 0 {
		return domain.EmbeddingResponse{}, domain.ErrInvalidRequest
	}

	scenario, err := e.Scenarios.Resolve(ctx, domain.ScenarioKey{Model: req.Model, Operation: "embedding"})
	if err != nil {
		return domain.EmbeddingResponse{}, err
	}
	if scenario.Error != nil {
		return domain.EmbeddingResponse{}, &domain.InjectedError{Injection: *scenario.Error}
	}

	dims := e.Dims
	if dims <= 0 {
		dims = 8
	}

	embeddings := make([]domain.Embedding, len(req.Input))
	promptTokens := 0
	for i, in := range req.Input {
		embeddings[i] = domain.Embedding{Index: i, Vector: deterministicVector(in, dims), InputText: in}
		promptTokens += len(in) / 4
	}

	return domain.EmbeddingResponse{
		Model:      req.Model,
		Embeddings: embeddings,
		Usage:      domain.TokenUsage{PromptTokens: promptTokens, TotalTokens: promptTokens},
	}, nil
}

// deterministicVector derives a stable pseudo-embedding from the input text so repeated calls
// with the same input are reproducible in tests, without needing any real model.
func deterministicVector(text string, dims int) []float64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(text))
	seed := h.Sum64()

	vec := make([]float64, dims)
	for i := range vec {
		seed = seed*6364136223846793005 + 1442695040888963407
		vec[i] = float64(int64(seed)%1000) / 1000.0
	}
	return vec
}
