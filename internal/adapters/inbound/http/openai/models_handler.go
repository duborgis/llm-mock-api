package openai

import (
	"encoding/json"
	"net/http"

	oai "github.com/openai/openai-go"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func toSDKModel(m domain.ModelInfo) oai.Model {
	return oai.Model{ID: m.ID, Created: m.Created, Object: "model", OwnedBy: m.Owner}
}

// modelListPage matches the wire shape of openai-go's pagination.Page[Model]: {"object":"list","data":[...]}.
type modelListPage struct {
	Object string      `json:"object"`
	Data   []oai.Model `json:"data"`
}

func (h *Handler) listModels(w http.ResponseWriter, r *http.Request) {
	models, err := h.Models.List(r.Context())
	if err != nil {
		writeError(w, h.Logger, err)
		return
	}
	data := make([]oai.Model, len(models))
	for i, m := range models {
		data[i] = toSDKModel(m)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(modelListPage{Object: "list", Data: data})
}

func (h *Handler) getModel(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	m, err := h.Models.Get(r.Context(), id)
	if err != nil {
		writeError(w, h.Logger, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toSDKModel(m))
}
