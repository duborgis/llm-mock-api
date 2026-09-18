package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/responses"
	"github.com/openai/openai-go/shared/constant"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// The Responses API (POST /v1/responses) reuses the same generic ChatCompletionUseCase as
// /v1/chat/completions — it's the same underlying capability (send messages, get a model
// reply) with a different wire envelope. Only this file's translation to/from
// responses.Response and responses.ResponseNewParams (openai-go's own structs) differs.

// toDomainRequestFromResponses extracts a generic domain.ChatRequest from an SDK
// ResponseNewParams. Supports the plain string Input and the input-items-list Input
// (EasyInputMessage variant, which covers role+string-content messages — the common case).
func toDomainRequestFromResponses(raw []byte) (responses.ResponseNewParams, domain.ChatRequest, error) {
	var params responses.ResponseNewParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return params, domain.ChatRequest{}, fmt.Errorf("%w: %s", domain.ErrInvalidRequest, err)
	}
	if params.Model == "" {
		return params, domain.ChatRequest{}, domain.ErrInvalidRequest
	}

	req := domain.ChatRequest{Model: string(params.Model)}
	if params.Temperature.Valid() {
		req.Temperature = params.Temperature.Value
	}
	if params.MaxOutputTokens.Valid() {
		req.MaxTokens = int(params.MaxOutputTokens.Value)
	}
	if params.Instructions.Valid() && params.Instructions.Value != "" {
		req.Messages = append(req.Messages, domain.ChatMessage{
			Role:    domain.RoleSystem,
			Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: params.Instructions.Value}},
		})
	}

	switch {
	case !param.IsOmitted(params.Input.OfString):
		req.Messages = append(req.Messages, domain.ChatMessage{
			Role:    domain.RoleUser,
			Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: params.Input.OfString.Value}},
		})
	case !param.IsOmitted(params.Input.OfInputItemList):
		for _, item := range params.Input.OfInputItemList {
			role := domain.RoleUser
			if r := item.GetRole(); r != nil {
				role = domain.Role(*r)
			}
			text := ""
			if content := item.GetContent(); content.AsAny() != nil {
				if s, ok := content.AsAny().(*string); ok && s != nil {
					text = *s
				}
			}
			req.Messages = append(req.Messages, domain.ChatMessage{
				Role:    role,
				Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: text}},
			})
		}
	}

	if len(req.Messages) == 0 {
		return params, domain.ChatRequest{}, domain.ErrInvalidRequest
	}
	return params, req, nil
}

// toSDKResponse builds a responses.Response (openai-go's own struct) from a generic
// domain.ChatResponse, so the JSON shape is exactly what openai-go's response decoder expects.
func toSDKResponse(resp domain.ChatResponse, stream bool) responses.Response {
	var content []responses.ResponseOutputMessageContentUnion
	var outputText string
	if len(resp.Choices) > 0 {
		outputText = resp.Choices[0].Message.Text()
	}
	content = append(content, responses.ResponseOutputMessageContentUnion{
		Type:        "output_text",
		Text:        outputText,
		Annotations: []responses.ResponseOutputTextAnnotationUnion{},
	})

	return responses.Response{
		ID:        resp.ID,
		CreatedAt: float64(resp.Created),
		Model:     resp.Model,
		Object:    constant.Response("response"),
		Status:    responses.ResponseStatusCompleted,
		Output: []responses.ResponseOutputItemUnion{{
			ID:      resp.ID + "-msg",
			Type:    "message",
			Role:    constant.Assistant("assistant"),
			Status:  "completed",
			Content: content,
		}},
		Usage: responses.ResponseUsage{
			InputTokens:  int64(resp.Usage.PromptTokens),
			OutputTokens: int64(resp.Usage.CompletionTokens),
			TotalTokens:  int64(resp.Usage.TotalTokens),
		},
	}
}

func (h *Handler) responses(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, h.Logger, domain.ErrInvalidRequest)
		return
	}

	params, req, err := toDomainRequestFromResponses(raw)
	if err != nil {
		if h.Logger != nil {
			h.Logger.Error("openai: failed to decode responses request body", "err", err, "body", string(raw))
		}
		writeError(w, h.Logger, err)
		return
	}

	var streamPeek struct {
		Stream bool `json:"stream"`
	}
	_ = json.Unmarshal(raw, &streamPeek)

	if streamPeek.Stream {
		h.streamResponses(w, r, req)
		return
	}

	resp, err := h.Chat.Complete(r.Context(), req)
	if err != nil {
		writeError(w, h.Logger, err)
		return
	}

	_ = params // decoded for validation/field extraction above; not needed further for the mock
	out := toSDKResponse(resp, false)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// streamResponses writes the Responses API's SSE framing: `event: <type>\ndata: {...}\n\n`
// per event. openai-go's ssestream decoder unmarshals each `data:` payload directly into the
// union type by its own "type" field, so each event's JSON must be self-describing.
func (h *Handler) streamResponses(w http.ResponseWriter, r *http.Request, req domain.ChatRequest) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, h.Logger, fmt.Errorf("streaming unsupported"))
		return
	}

	ch, err := h.Chat.Stream(r.Context(), req)
	if err != nil {
		writeError(w, h.Logger, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	writeEvent := func(eventType string, payload any) {
		data, _ := json.Marshal(payload)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
		flusher.Flush()
	}

	id := fmt.Sprintf("resp-%d", 0)
	seq := int64(0)
	nextSeq := func() int64 { seq++; return seq }

	itemID := ""
	var fullText string
	for chunk := range ch {
		if id == "resp-0" {
			id = chunk.ID
			writeEvent("response.created", responses.ResponseCreatedEvent{
				Type:           constant.ResponseCreated("response.created"),
				SequenceNumber: nextSeq(),
				Response:       toSDKResponse(domain.ChatResponse{ID: id, Model: chunk.Model, Created: chunk.Created}, true),
			})
			itemID = id + "-msg"
			writeEvent("response.output_item.added", responses.ResponseOutputItemAddedEvent{
				Type:           constant.ResponseOutputItemAdded("response.output_item.added"),
				SequenceNumber: nextSeq(),
				OutputIndex:    0,
				Item: responses.ResponseOutputItemUnion{
					ID: itemID, Type: "message", Role: constant.Assistant("assistant"), Status: "in_progress",
				},
			})
		}

		fullText += chunk.DeltaContent
		if chunk.DeltaContent != "" {
			writeEvent("response.output_text.delta", responses.ResponseTextDeltaEvent{
				Type:           constant.ResponseOutputTextDelta("response.output_text.delta"),
				SequenceNumber: nextSeq(),
				ItemID:         itemID,
				OutputIndex:    0,
				ContentIndex:   0,
				Delta:          chunk.DeltaContent,
				Logprobs:       []responses.ResponseTextDeltaEventLogprob{},
			})
		}

		if chunk.FinishReason != "" {
			finalResp := toSDKResponse(domain.ChatResponse{
				ID: id, Model: chunk.Model, Created: chunk.Created,
				Choices: []domain.Choice{{Message: domain.ChatMessage{
					Role: domain.RoleAssistant, Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: fullText}},
				}, FinishReason: chunk.FinishReason}},
				Usage: usageOrZero(chunk),
			}, false)
			writeEvent("response.output_item.done", responses.ResponseOutputItemDoneEvent{
				Type:           constant.ResponseOutputItemDone("response.output_item.done"),
				SequenceNumber: nextSeq(),
				OutputIndex:    0,
				Item:           finalResp.Output[0],
			})
			writeEvent("response.completed", responses.ResponseCompletedEvent{
				Type:           constant.ResponseCompleted("response.completed"),
				SequenceNumber: nextSeq(),
				Response:       finalResp,
			})
		}
	}
}

func usageOrZero(chunk domain.StreamChunk) domain.TokenUsage {
	if chunk.Usage != nil {
		return *chunk.Usage
	}
	return domain.TokenUsage{}
}
