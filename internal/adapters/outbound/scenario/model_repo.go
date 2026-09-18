package scenario

import (
	"context"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// ModelRepo implements ports.ModelRepository as a static in-memory catalog covering
// representative models for OpenAI, Bedrock, and Vertex, so all three inbound adapters'
// model-listing endpoints have something faithful to return.
type ModelRepo struct {
	models []domain.ModelInfo
}

func NewModelRepo() *ModelRepo {
	return &ModelRepo{models: []domain.ModelInfo{
		// --- OpenAI native API (also listed as-is; Bedrock exposes OpenAI's *open-weight*
		// gpt-oss models separately below, since that's the only OpenAI family AWS actually hosts) ---
		{ID: "gpt-5", Owner: "openai", Created: 1755000000, Family: "gpt", Providers: []string{"openai"}},
		{ID: "gpt-5-mini", Owner: "openai", Created: 1755000000, Family: "gpt", Providers: []string{"openai"}},
		{ID: "gpt-5-nano", Owner: "openai", Created: 1755000000, Family: "gpt", Providers: []string{"openai"}},
		{ID: "gpt-4.1", Owner: "openai", Created: 1744000000, Family: "gpt", Providers: []string{"openai"}},
		{ID: "gpt-4.1-mini", Owner: "openai", Created: 1744000000, Family: "gpt", Providers: []string{"openai"}},
		{ID: "gpt-4o", Owner: "openai", Created: 1715367049, Family: "gpt", Providers: []string{"openai"}},
		{ID: "gpt-4o-mini", Owner: "openai", Created: 1721172741, Family: "gpt", Providers: []string{"openai"}},
		{ID: "o3", Owner: "openai", Created: 1744000000, Family: "gpt", Providers: []string{"openai"}},
		{ID: "o4-mini", Owner: "openai", Created: 1744000000, Family: "gpt", Providers: []string{"openai"}},
		{ID: "text-embedding-3-small", Owner: "openai", Created: 1705948997, Family: "embedding", Providers: []string{"openai"}},
		{ID: "text-embedding-3-large", Owner: "openai", Created: 1705948997, Family: "embedding", Providers: []string{"openai"}},
		{ID: "dall-e-3", Owner: "openai", Created: 1698785189, Family: "image", Providers: []string{"openai"}},
		{ID: "gpt-image-1", Owner: "openai", Created: 1745000000, Family: "image", Providers: []string{"openai"}},
		{ID: "whisper-1", Owner: "openai", Created: 1677532384, Family: "audio", Providers: []string{"openai"}},
		{ID: "gpt-4o-mini-transcribe", Owner: "openai", Created: 1741000000, Family: "audio", Providers: []string{"openai"}},
		{ID: "tts-1", Owner: "openai", Created: 1699000000, Family: "audio", Providers: []string{"openai"}},
		{ID: "gpt-4o-mini-tts", Owner: "openai", Created: 1741000000, Family: "audio", Providers: []string{"openai"}},
		{ID: "sora-2", Owner: "openai", Created: 1759000000, Family: "video", Providers: []string{"openai"}},
		{ID: "sora-2-pro", Owner: "openai", Created: 1759000000, Family: "video", Providers: []string{"openai"}},

		// --- OpenAI open-weight models, hosted on Bedrock (AWS-side, not the OpenAI API surface) ---
		{ID: "openai.gpt-oss-120b-1:0", Owner: "openai", Created: 1754000000, Family: "gpt-oss", Providers: []string{"bedrock"}},
		{ID: "openai.gpt-oss-20b-1:0", Owner: "openai", Created: 1754000000, Family: "gpt-oss", Providers: []string{"bedrock"}},

		// --- Anthropic Claude, on Bedrock ---
		{ID: "anthropic.claude-opus-4-5-20251101-v1:0", Owner: "anthropic", Created: 1762000000, Family: "claude", Providers: []string{"bedrock"}},
		{ID: "anthropic.claude-sonnet-4-5-20250929-v1:0", Owner: "anthropic", Created: 1759000000, Family: "claude", Providers: []string{"bedrock"}},
		{ID: "anthropic.claude-haiku-4-5-20251001-v1:0", Owner: "anthropic", Created: 1759349000, Family: "claude", Providers: []string{"bedrock"}},
		{ID: "anthropic.claude-3-5-sonnet-20241022-v2:0", Owner: "anthropic", Created: 1729555200, Family: "claude", Providers: []string{"bedrock"}},

		// --- Amazon Nova, on Bedrock ---
		{ID: "amazon.nova-pro-v1:0", Owner: "amazon", Created: 1733011200, Family: "nova", Providers: []string{"bedrock"}},
		{ID: "amazon.nova-lite-v1:0", Owner: "amazon", Created: 1733011200, Family: "nova", Providers: []string{"bedrock"}},
		{ID: "amazon.nova-micro-v1:0", Owner: "amazon", Created: 1733011200, Family: "nova", Providers: []string{"bedrock"}},
		{ID: "amazon.titan-embed-text-v2:0", Owner: "amazon", Created: 1707350400, Family: "embedding", Providers: []string{"bedrock"}},

		// --- Meta Llama, on Bedrock ---
		{ID: "meta.llama4-maverick-17b-instruct-v1:0", Owner: "meta", Created: 1743984000, Family: "llama", Providers: []string{"bedrock"}},
		{ID: "meta.llama4-scout-17b-instruct-v1:0", Owner: "meta", Created: 1743984000, Family: "llama", Providers: []string{"bedrock"}},
		{ID: "meta.llama3-3-70b-instruct-v1:0", Owner: "meta", Created: 1733788800, Family: "llama", Providers: []string{"bedrock"}},

		// --- Mistral, on Bedrock ---
		{ID: "mistral.mistral-large-2407-v1:0", Owner: "mistral", Created: 1720915200, Family: "mistral", Providers: []string{"bedrock"}},

		// --- Google Gemini, on Vertex AI ---
		{ID: "gemini-2.5-pro", Owner: "google", Created: 1750000000, Family: "gemini", Providers: []string{"vertex"}},
		{ID: "gemini-2.5-flash", Owner: "google", Created: 1750000000, Family: "gemini", Providers: []string{"vertex"}},
		{ID: "gemini-2.5-flash-lite", Owner: "google", Created: 1750000000, Family: "gemini", Providers: []string{"vertex"}},
		{ID: "gemini-2.0-flash", Owner: "google", Created: 1738800000, Family: "gemini", Providers: []string{"vertex"}},
		{ID: "text-embedding-005", Owner: "google", Created: 1715644800, Family: "embedding", Providers: []string{"vertex"}},
	}}
}

func (r *ModelRepo) List(ctx context.Context) ([]domain.ModelInfo, error) {
	return r.models, nil
}

func (r *ModelRepo) Get(ctx context.Context, id string) (domain.ModelInfo, error) {
	for _, m := range r.models {
		if m.ID == id {
			return m, nil
		}
	}
	return domain.ModelInfo{}, domain.ErrModelNotFound
}
