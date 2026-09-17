package bedrock

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// converseStream event payload shapes (Bedrock ConverseStream event stream), keyed by
// :event-type header value: messageStart, contentBlockDelta, contentBlockStop, messageStop, metadata.
type csMessageStart struct {
	Role string `json:"role"`
}
type csDelta struct {
	Text string `json:"text"`
}
type csContentBlockDelta struct {
	ContentBlockIndex int     `json:"contentBlockIndex"`
	Delta             csDelta `json:"delta"`
}
type csContentBlockStop struct {
	ContentBlockIndex int `json:"contentBlockIndex"`
}
type csMessageStop struct {
	StopReason string `json:"stopReason"`
}
type csMetadata struct {
	Usage   converseUsage   `json:"usage"`
	Metrics converseMetrics `json:"metrics"`
}

func (h *Handler) converseStream(w http.ResponseWriter, r *http.Request) {
	modelID := r.PathValue("modelId")

	var body converseRequestBody
	if err := readJSON(r.Body, &body); err != nil || len(body.Messages) == 0 {
		writeBedrockError(w, domain.ErrInvalidRequest)
		return
	}
	req := toDomainRequest(modelID, body.Messages, body.System)

	ch, err := h.Chat.Stream(r.Context(), req)
	if err != nil {
		writeBedrockError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
	w.WriteHeader(http.StatusOK)
	enc := eventstream.NewEncoder()

	startPayload, _ := json.Marshal(csMessageStart{Role: "assistant"})
	_ = writeEventStreamMessage(w, enc, "messageStart", startPayload)

	var finish domain.FinishReason = domain.FinishStop
	var usage domain.TokenUsage
	for chunk := range ch {
		if chunk.DeltaContent != "" {
			p, _ := json.Marshal(csContentBlockDelta{ContentBlockIndex: 0, Delta: csDelta{Text: chunk.DeltaContent}})
			_ = writeEventStreamMessage(w, enc, "contentBlockDelta", p)
		}
		if chunk.FinishReason != "" {
			finish = chunk.FinishReason
		}
		if chunk.Usage != nil {
			usage = *chunk.Usage
		}
	}

	stopPayload, _ := json.Marshal(csContentBlockStop{ContentBlockIndex: 0})
	_ = writeEventStreamMessage(w, enc, "contentBlockStop", stopPayload)

	msgStopPayload, _ := json.Marshal(csMessageStop{StopReason: stopReasonFor(finish)})
	_ = writeEventStreamMessage(w, enc, "messageStop", msgStopPayload)

	metaPayload, _ := json.Marshal(csMetadata{
		Usage: converseUsage{
			InputTokens:  usage.PromptTokens,
			OutputTokens: usage.CompletionTokens,
			TotalTokens:  usage.TotalTokens,
		},
		Metrics: converseMetrics{LatencyMs: 1},
	})
	_ = writeEventStreamMessage(w, enc, "metadata", metaPayload)
}
