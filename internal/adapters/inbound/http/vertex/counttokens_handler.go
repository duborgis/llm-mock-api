package vertex

import (
	"encoding/json"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func (h *Handler) countTokens(w http.ResponseWriter, r *http.Request, model string) {
	var body countTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Contents) == 0 {
		writeVertexError(w, domain.ErrInvalidRequest)
		return
	}
	total := 0
	for _, c := range body.Contents {
		total += len(partsText(c.Parts)) / 4
	}
	if total == 0 {
		total = 1
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(countTokensResponse{TotalTokens: int32(total)})
}
