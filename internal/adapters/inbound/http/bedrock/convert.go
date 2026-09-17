package bedrock

import (
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func toDomainRequest(modelID string, msgs []converseMessage, system []converseContentBlock) domain.ChatRequest {
	req := domain.ChatRequest{Model: modelID}
	for _, sb := range system {
		if sb.Text != "" {
			req.Messages = append(req.Messages, domain.ChatMessage{
				Role:    domain.RoleSystem,
				Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: sb.Text}},
			})
		}
	}
	for _, m := range msgs {
		text := ""
		for _, c := range m.Content {
			text += c.Text
		}
		req.Messages = append(req.Messages, domain.ChatMessage{
			Role:    domain.Role(m.Role),
			Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: text}},
		})
	}
	return req
}

func stopReasonFor(f domain.FinishReason) string {
	switch f {
	case domain.FinishLength:
		return "max_tokens"
	case domain.FinishContentFilter:
		return "content_filtered"
	default:
		return "end_turn"
	}
}

func writeBedrockError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	msg := err.Error()

	if ie, ok := err.(*domain.InjectedError); ok {
		status = ie.Injection.StatusCode
		if status == 0 {
			status = http.StatusInternalServerError
		}
		msg = ie.Injection.Message
	} else if err == domain.ErrInvalidRequest {
		status = http.StatusBadRequest
	} else if err == domain.ErrModelNotFound {
		status = http.StatusNotFound
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = writeJSON(w, converseErrorBody{Message: msg})
}
