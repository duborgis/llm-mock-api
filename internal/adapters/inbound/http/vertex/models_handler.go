package vertex

import (
	"encoding/json"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// vertexModel mirrors Vertex's Model resource wire shape (minimal subset).
type vertexModel struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type listModelsResponse struct {
	Models []vertexModel `json:"models"`
}

func toVertexModel(m domain.ModelInfo) vertexModel {
	return vertexModel{
		Name:        "publishers/google/models/" + m.ID,
		DisplayName: m.ID,
	}
}

func (h *Handler) listModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.Models.List(r.Context())
	if err != nil {
		writeVertexError(w, err)
		return
	}
	out := listModelsResponse{}
	for _, m := range models {
		if m.HasProvider("vertex") {
			out.Models = append(out.Models, toVertexModel(m))
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
