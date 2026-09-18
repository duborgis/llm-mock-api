package ports

import (
	"context"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// ChatCompletionUseCase is the driving port for chat/message completion, implemented by usecase.Chat
// and called by inbound HTTP adapters (openai/bedrock/vertex routers).
type ChatCompletionUseCase interface {
	Complete(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error)
	Stream(ctx context.Context, req domain.ChatRequest) (<-chan domain.StreamChunk, error)
}

// EmbeddingUseCase is the driving port for embedding generation.
type EmbeddingUseCase interface {
	Embed(ctx context.Context, req domain.EmbeddingRequest) (domain.EmbeddingResponse, error)
}

// ImageGenerationUseCase is the driving port for text-to-image generation.
type ImageGenerationUseCase interface {
	Generate(ctx context.Context, req domain.ImageRequest) (domain.ImageResponse, error)
}

// ModelCatalogUseCase is the driving port for model listing/lookup.
type ModelCatalogUseCase interface {
	List(ctx context.Context) ([]domain.ModelInfo, error)
	Get(ctx context.Context, id string) (domain.ModelInfo, error)
}
