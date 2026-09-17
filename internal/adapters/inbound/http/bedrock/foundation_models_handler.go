package bedrock

import (
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// foundationModelSummary mirrors bedrock (control plane)'s ListFoundationModels wire shape.
type foundationModelSummary struct {
	ModelID           string `json:"modelId"`
	ModelName         string `json:"modelName"`
	ProviderName      string `json:"providerName"`
	ResponseStreaming bool   `json:"responseStreamingSupported"`
}

type listFoundationModelsResponse struct {
	ModelSummaries []foundationModelSummary `json:"modelSummaries"`
}

func toFoundationModel(m domain.ModelInfo) foundationModelSummary {
	return foundationModelSummary{
		ModelID:           m.ID,
		ModelName:         m.ID,
		ProviderName:      m.Owner,
		ResponseStreaming: true,
	}
}

func (h *Handler) listFoundationModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.Models.List(r.Context())
	if err != nil {
		writeBedrockError(w, err)
		return
	}
	out := listFoundationModelsResponse{}
	for _, m := range models {
		out.ModelSummaries = append(out.ModelSummaries, toFoundationModel(m))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = writeJSON(w, out)
}

func (h *Handler) getFoundationModel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("modelId")
	m, err := h.Models.Get(r.Context(), id)
	if err != nil {
		writeBedrockError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = writeJSON(w, struct {
		ModelDetails foundationModelSummary `json:"modelDetails"`
	}{ModelDetails: toFoundationModel(m)})
}
