package openai

import (
	"encoding/json"
	"net/http"

	oai "github.com/openai/openai-go"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func (h *Handler) createImage(w http.ResponseWriter, r *http.Request) {
	var params oai.ImageGenerateParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, h.Logger, domain.ErrInvalidRequest)
		return
	}
	if params.Prompt == "" {
		writeError(w, h.Logger, domain.ErrInvalidRequest)
		return
	}

	model := string(params.Model)
	if model == "" {
		model = string(oai.ImageModelDallE2)
	}

	resp, err := h.Images.Generate(r.Context(), domain.ImageRequest{
		Model:  model,
		Prompt: params.Prompt,
		N:      int(params.N.Value),
		Size:   string(params.Size),
		User:   params.User.Value,
	})
	if err != nil {
		writeError(w, h.Logger, err)
		return
	}

	data := make([]oai.Image, len(resp.Images))
	for i, img := range resp.Images {
		data[i] = oai.Image{B64JSON: img.B64JSON, RevisedPrompt: img.RevisedPrompt}
	}
	out := oai.ImagesResponse{
		Created: resp.Created,
		Data:    data,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
