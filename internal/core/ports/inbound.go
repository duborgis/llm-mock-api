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

// AudioTranscriptionUseCase is the driving port for speech-to-text.
type AudioTranscriptionUseCase interface {
	Transcribe(ctx context.Context, req domain.TranscriptionRequest) (domain.TranscriptionResponse, error)
}

// AudioSpeechUseCase is the driving port for text-to-speech.
type AudioSpeechUseCase interface {
	Synthesize(ctx context.Context, req domain.SpeechRequest) (domain.SpeechResponse, error)
}

// VideoGenerationUseCase is the driving port for Sora-style video generation. Modeled as a
// job (Create/Get/Content) rather than a single call, since that's how the real API works —
// see domain.Video.
type VideoGenerationUseCase interface {
	Create(ctx context.Context, req domain.VideoRequest) (domain.Video, error)
	Get(ctx context.Context, id string) (domain.Video, error)
	Content(ctx context.Context, id string) ([]byte, string, error)
}

// ModelCatalogUseCase is the driving port for model listing/lookup.
type ModelCatalogUseCase interface {
	List(ctx context.Context) ([]domain.ModelInfo, error)
	Get(ctx context.Context, id string) (domain.ModelInfo, error)
}
