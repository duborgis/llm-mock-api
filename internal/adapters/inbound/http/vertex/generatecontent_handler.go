package vertex

import (
	"encoding/json"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func (h *Handler) generateContent(w http.ResponseWriter, r *http.Request, model string) {
	var body generateContentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Contents) == 0 {
		writeVertexError(w, domain.ErrInvalidRequest)
		return
	}

	req := toDomainRequest(model, body.Contents, body.SystemInstruction)
	resp, err := h.Chat.Complete(r.Context(), req)
	if err != nil {
		writeVertexError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toGenaiResponse(resp))
}
