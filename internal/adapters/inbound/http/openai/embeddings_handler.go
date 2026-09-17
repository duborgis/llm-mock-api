package openai

import (
	"encoding/json"
	"net/http"

	oai "github.com/openai/openai-go"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func (h *Handler) createEmbeddings(w http.ResponseWriter, r *http.Request) {
	var params oai.EmbeddingNewParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, domain.ErrInvalidRequest)
		return
	}

	var inputs []string
	if params.Input.OfString.Valid() {
		inputs = []string{params.Input.OfString.Value}
	} else if len(params.Input.OfArrayOfStrings) > 0 {
		inputs = params.Input.OfArrayOfStrings
	}
	if params.Model == "" || len(inputs) == 0 {
		writeError(w, domain.ErrInvalidRequest)
		return
	}

	resp, err := h.Embed.Embed(r.Context(), domain.EmbeddingRequest{Model: string(params.Model), Input: inputs})
	if err != nil {
		writeError(w, err)
		return
	}

	data := make([]oai.Embedding, len(resp.Embeddings))
	for i, e := range resp.Embeddings {
		data[i] = oai.Embedding{Embedding: e.Vector, Index: int64(e.Index), Object: "embedding"}
	}
	out := oai.CreateEmbeddingResponse{
		Data:   data,
		Model:  resp.Model,
		Object: "list",
		Usage: oai.CreateEmbeddingResponseUsage{
			PromptTokens: int64(resp.Usage.PromptTokens),
			TotalTokens:  int64(resp.Usage.TotalTokens),
		},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
