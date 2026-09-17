package bedrock

import (
	"fmt"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// invokeModel implements POST /model/{modelId}/invoke for the Anthropic Claude-on-Bedrock body
// shape (the most common Bedrock invoke case). InvokeModelInput/Output's Body field is an
// opaque []byte as far as the SDK/smithy protocol are concerned, so what we write here only
// needs to match Anthropic's documented Messages body shape — which it does, field for field.
func (h *Handler) invokeModel(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("modelId")

	var body claudeInvokeRequestBody
	if err := readJSON(r.Body, &body); err != nil || len(body.Messages) == 0 {
		writeBedrockError(w, domain.ErrInvalidRequest)
		return
	}

	req := domain.ChatRequest{Model: modelID, MaxTokens: body.MaxTokens, Temperature: body.Temperature}
	if body.System != "" {
		req.Messages = append(req.Messages, domain.ChatMessage{
			Role:    domain.RoleSystem,
			Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: body.System}},
		})
	}
	for _, m := range body.Messages {
		text := ""
		for _, c := range m.Content {
			text += c.Text
		}
		req.Messages = append(req.Messages, domain.ChatMessage{
			Role:    domain.Role(m.Role),
			Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: text}},
		})
	}

	resp, err := h.Chat.Complete(r.Context(), req)
	if err != nil {
		writeBedrockError(w, err)
		return
	}
	choice := resp.Choices[0]

	out := claudeInvokeResponseBody{
		ID:    fmt.Sprintf("msg_%s", resp.ID),
		Type:  "message",
		Role:  "assistant",
		Model: modelID,
		Content: []claudeInvokeContentBlock{
			{Type: "text", Text: choice.Message.Text()},
		},
		StopReason: claudeStopReasonFor(choice.FinishReason),
		Usage: claudeInvokeUsage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = writeJSON(w, out)
}

func claudeStopReasonFor(f domain.FinishReason) string {
	switch f {
	case domain.FinishLength:
		return "max_tokens"
	case domain.FinishContentFilter:
		return "stop_sequence"
	default:
		return "end_turn"
	}
}
