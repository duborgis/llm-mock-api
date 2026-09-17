package vertex

import (
	"encoding/json"
	"net/http"

	"google.golang.org/genai"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

func contentToDomainRole(role string) domain.Role {
	if role == "model" {
		return domain.RoleAssistant
	}
	return domain.RoleUser
}

func toDomainRequest(model string, contents []*genai.Content, system *genai.Content) domain.ChatRequest {
	req := domain.ChatRequest{Model: model}
	if system != nil {
		req.Messages = append(req.Messages, domain.ChatMessage{
			Role:    domain.RoleSystem,
			Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: partsText(system.Parts)}},
		})
	}
	for _, c := range contents {
		req.Messages = append(req.Messages, domain.ChatMessage{
			Role:    contentToDomainRole(c.Role),
			Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: partsText(c.Parts)}},
		})
	}
	return req
}

func partsText(parts []*genai.Part) string {
	out := ""
	for _, p := range parts {
		if p != nil {
			out += p.Text
		}
	}
	return out
}

func finishReasonFor(f domain.FinishReason) genai.FinishReason {
	switch f {
	case domain.FinishLength:
		return genai.FinishReasonMaxTokens
	case domain.FinishContentFilter:
		return genai.FinishReasonSafety
	default:
		return genai.FinishReasonStop
	}
}

func toGenaiResponse(resp domain.ChatResponse) *genai.GenerateContentResponse {
	candidates := make([]*genai.Candidate, len(resp.Choices))
	for i, c := range resp.Choices {
		candidates[i] = &genai.Candidate{
			Content: &genai.Content{
				Role:  "model",
				Parts: []*genai.Part{{Text: c.Message.Text()}},
			},
			FinishReason: finishReasonFor(c.FinishReason),
			Index:        int32(c.Index),
		}
	}
	return &genai.GenerateContentResponse{
		Candidates:   candidates,
		ModelVersion: resp.Model,
		ResponseID:   resp.ID,
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
			PromptTokenCount:     int32(resp.Usage.PromptTokens),
			CandidatesTokenCount: int32(resp.Usage.CompletionTokens),
			TotalTokenCount:      int32(resp.Usage.TotalTokens),
		},
	}
}

func writeVertexError(w http.ResponseWriter, err error) {
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
	body := struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}{}
	body.Error.Code = status
	body.Error.Message = msg
	body.Error.Status = "INTERNAL"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
