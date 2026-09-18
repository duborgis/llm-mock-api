package usecase

import (
	"context"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
	"github.com/duborgis/llm-mock-api/internal/core/ports"
)

// tiny1x1PNG is a real, valid 1x1 transparent PNG. The mock never renders actual pixels —
// this placeholder lets callers round-trip a syntactically valid b64_json image payload.
const tiny1x1PNG = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="

// Image implements ports.ImageGenerationUseCase.
type Image struct {
	Scenarios ports.ScenarioRepository
	Clock     ports.Clock
}

func NewImage(scenarios ports.ScenarioRepository, clock ports.Clock) *Image {
	return &Image{Scenarios: scenarios, Clock: clock}
}

func (i *Image) Generate(ctx context.Context, req domain.ImageRequest) (domain.ImageResponse, error) {
	if req.Model == "" || req.Prompt == "" {
		return domain.ImageResponse{}, domain.ErrInvalidRequest
	}

	scenario, err := i.Scenarios.Resolve(ctx, domain.ScenarioKey{Model: req.Model, Operation: "image"})
	if err != nil {
		return domain.ImageResponse{}, err
	}
	if scenario.Error != nil {
		return domain.ImageResponse{}, &domain.InjectedError{Injection: *scenario.Error}
	}

	n := req.N
	if n <= 0 {
		n = 1
	}

	images := make([]domain.GeneratedImage, n)
	for idx := range images {
		images[idx] = domain.GeneratedImage{B64JSON: tiny1x1PNG, RevisedPrompt: req.Prompt}
	}

	return domain.ImageResponse{
		Created: i.Clock.Now().Unix(),
		Images:  images,
		Usage:   domain.TokenUsage{PromptTokens: len(req.Prompt) / 4, TotalTokens: len(req.Prompt) / 4},
	}, nil
}
