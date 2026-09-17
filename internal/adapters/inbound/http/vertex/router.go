// Package vertex is the driving HTTP adapter that speaks Vertex AI's Gemini REST contract
// (generateContent / streamGenerateContent / countTokens). Unlike the Bedrock SDK, google's
// genai package types (Content, Part, Candidate, GenerateContentResponse, ...) DO carry
// standard encoding/json tags, so we use them directly for response encoding — fidelity is
// structural. Vertex's Go SDK support for custom base URLs is limited, so
// test/contract/vertex_contract_test.go uses a plain http.Client against this server and
// unmarshals into these same genai structs, which is the fidelity property that matters.
package vertex

import (
	"log/slog"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/ports"
)

type Handler struct {
	Chat   ports.ChatCompletionUseCase
	Models ports.ModelCatalogUseCase
	Logger *slog.Logger
}

func NewRouter(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /v1/projects/{project}/locations/{location}/publishers/google/models/{model}", h.dispatchColonAction)
	mux.HandleFunc("GET /v1/publishers/google/models", h.listModels)
}

// dispatchColonAction handles Vertex's `:action` suffix routing (generateContent,
// streamGenerateContent, countTokens), which Go's net/http ServeMux pattern matching can't
// express directly since the action is appended to the path segment after a literal colon
// (e.g. ".../models/gemini-1.5-pro:generateContent").
func (h *Handler) dispatchColonAction(w http.ResponseWriter, r *http.Request) {
	model := r.PathValue("model")
	action := ""
	for i := len(model) - 1; i >= 0; i-- {
		if model[i] == ':' {
			action = model[i+1:]
			model = model[:i]
			break
		}
	}
	switch action {
	case "generateContent":
		h.generateContent(w, r, model)
	case "streamGenerateContent":
		h.streamGenerateContent(w, r, model)
	case "countTokens":
		h.countTokens(w, r, model)
	default:
		http.NotFound(w, r)
	}
}
