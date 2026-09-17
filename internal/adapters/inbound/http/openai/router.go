// Package openai is the driving HTTP adapter that speaks OpenAI's wire contract.
// It decodes/encodes using openai-go's own request/response struct types so fidelity
// is structural: if openai-go can round-trip a payload, the mock is faithful.
package openai

import (
	"log/slog"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/ports"
)

// Handler wires the OpenAI-shaped routes to the generic use cases.
type Handler struct {
	Chat   ports.ChatCompletionUseCase
	Embed  ports.EmbeddingUseCase
	Models ports.ModelCatalogUseCase
	Logger *slog.Logger
}

// NewRouter registers OpenAI-compatible routes onto mux.
func NewRouter(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /v1/chat/completions", h.chatCompletions)
	mux.HandleFunc("GET /v1/models", h.listModels)
	mux.HandleFunc("GET /v1/models/{id}", h.getModel)
	mux.HandleFunc("POST /v1/embeddings", h.createEmbeddings)
	mux.HandleFunc("POST /v1/moderations", h.moderations)
}
