package bedrock

import (
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func (h *Handler) converse(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("modelId")

	var body converseRequestBody
	if err := readJSON(r.Body, &body); err != nil || len(body.Messages) == 0 {
		writeBedrockError(w, domain.ErrInvalidRequest)
		return
	}

	req := toDomainRequest(modelID, body.Messages, body.System)
	resp, err := h.Chat.Complete(r.Context(), req)
	if err != nil {
		writeBedrockError(w, err)
		return
	}

	choice := resp.Choices[0]
	out := converseResponseBody{
		Output: converseOutputWrapper{Message: converseMessage{
			Role:    "assistant",
			Content: []converseContentBlock{{Text: choice.Message.Text()}},
		}},
		StopReason: stopReasonFor(choice.FinishReason),
		Usage: converseUsage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
		Metrics: converseMetrics{LatencyMs: 1},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = writeJSON(w, out)
}
