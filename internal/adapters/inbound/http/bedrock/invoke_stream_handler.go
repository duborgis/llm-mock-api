package bedrock

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// Claude-on-Bedrock streaming chunk shapes (Anthropic's streaming Messages API, forwarded
// verbatim as the "bytes" payload of each Bedrock "chunk" eventstream event).
type claudeStreamDelta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}
type claudeStreamEvent struct {
	Type       string             `json:"type"`
	Index      int                `json:"index,omitempty"`
	Delta      *claudeStreamDelta `json:"delta,omitempty"`
	Message    map[string]any     `json:"message,omitempty"`
	Usage      *claudeInvokeUsage `json:"usage,omitempty"`
	StopReason string             `json:"stop_reason,omitempty"`
}

// bedrockChunkPayload mirrors PayloadPart's wire JSON: {"bytes": base64(modelResponseChunkJSON)}.
type bedrockChunkPayload struct {
	Bytes []byte `json:"bytes"` // encoding/json base64-encodes []byte automatically
}

func (h *Handler) invokeModelStream(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("modelId")

	var body claudeInvokeRequestBody
	if err := readJSON(r.Body, &body); err != nil || len(body.Messages) == 0 {
		writeBedrockError(w, domain.ErrInvalidRequest)
		return
	}

	req := domain.ChatRequest{Model: modelID, MaxTokens: body.MaxTokens, Temperature: body.Temperature}
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

	ch, err := h.Chat.Stream(r.Context(), req)
	if err != nil {
		writeBedrockError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
	w.WriteHeader(http.StatusOK)
	enc := eventstream.NewEncoder()

	emit := func(ev claudeStreamEvent) {
		inner, _ := json.Marshal(ev)
		outer, _ := json.Marshal(bedrockChunkPayload{Bytes: inner})
		_ = writeEventStreamMessage(w, enc, "chunk", outer)
	}

	emit(claudeStreamEvent{Type: "message_start", Message: map[string]any{
		"id": fmt.Sprintf("msg_%s", modelID), "type": "message", "role": "assistant", "model": modelID,
	}})
	emit(claudeStreamEvent{Type: "content_block_start", Index: 0, Delta: &claudeStreamDelta{Type: "text", Text: ""}})

	var finish domain.FinishReason = domain.FinishStop
	var usage domain.TokenUsage
	for chunk := range ch {
		if chunk.DeltaContent != "" {
			emit(claudeStreamEvent{Type: "content_block_delta", Index: 0, Delta: &claudeStreamDelta{Type: "text_delta", Text: chunk.DeltaContent}})
		}
		if chunk.FinishReason != "" {
			finish = chunk.FinishReason
		}
		if chunk.Usage != nil {
			usage = *chunk.Usage
		}
	}

	emit(claudeStreamEvent{Type: "content_block_stop", Index: 0})
	emit(claudeStreamEvent{
		Type:       "message_delta",
		StopReason: claudeStopReasonFor(finish),
		Usage:      &claudeInvokeUsage{InputTokens: usage.PromptTokens, OutputTokens: usage.CompletionTokens},
	})
	emit(claudeStreamEvent{Type: "message_stop"})
}
