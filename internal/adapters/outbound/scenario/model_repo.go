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
		{ID: "gpt-4o", Owner: "openai", Created: 1715367049, Family: "gpt"},
		{ID: "gpt-4o-mini", Owner: "openai", Created: 1721172741, Family: "gpt"},
		{ID: "text-embedding-3-small", Owner: "openai", Created: 1705948997, Family: "embedding"},
		{ID: "anthropic.claude-3-5-sonnet-20240620-v1:0", Owner: "anthropic", Created: 1718841600, Family: "claude"},
		{ID: "amazon.titan-embed-text-v1", Owner: "amazon", Created: 1699920000, Family: "titan"},
		{ID: "meta.llama3-70b-instruct-v1:0", Owner: "meta", Created: 1713398400, Family: "llama"},
		{ID: "gemini-1.5-pro", Owner: "google", Created: 1715644800, Family: "gemini"},
		{ID: "gemini-1.5-flash", Owner: "google", Created: 1715644800, Family: "gemini"},
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
