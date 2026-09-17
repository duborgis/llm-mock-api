package openai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	oai "github.com/openai/openai-go"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// toDomainMessage extracts a generic domain.ChatMessage from an SDK message union, best-effort
// (only plain string content is fully supported, which covers the overwhelming common case).
func toDomainMessage(u oai.ChatCompletionMessageParamUnion) domain.ChatMessage {
	role := domain.RoleUser
	if r := u.GetRole(); r != nil {
		role = domain.Role(*r)
	}
	text := ""
	if content := u.GetContent(); content.AsAny() != nil {
		if s, ok := content.AsAny().(*string); ok && s != nil {
			text = *s
		}
	}
	return domain.ChatMessage{Role: role, Content: []domain.ContentPart{{Type: domain.ContentPartText, Text: text}}}
}

func (h *Handler) decodeChatRequest(r *http.Request) (oai.ChatCompletionNewParams, domain.ChatRequest, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return oai.ChatCompletionNewParams{}, domain.ChatRequest{}, domain.ErrInvalidRequest
	}

	var params oai.ChatCompletionNewParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return params, domain.ChatRequest{}, domain.ErrInvalidRequest
	}
	if params.Model == "" || len(params.Messages) == 0 {
		return params, domain.ChatRequest{}, domain.ErrInvalidRequest
	}

	// The SDK's ChatCompletionNewParams struct doesn't carry "stream" as a field (the client
	// injects it via option.WithJSONSet when calling the streaming method), so we peek at the
	// raw wire JSON for it directly.
	var streamPeek struct {
		Stream bool `json:"stream"`
	}
	_ = json.Unmarshal(raw, &streamPeek)

	req := domain.ChatRequest{
		Model:  string(params.Model),
		Stream: streamPeek.Stream,
	}
	if params.MaxTokens.Valid() {
		req.MaxTokens = int(params.MaxTokens.Value)
	}
	if params.Temperature.Valid() {
		req.Temperature = params.Temperature.Value
	}
	for _, m := range params.Messages {
		req.Messages = append(req.Messages, toDomainMessage(m))
	}
	return params, req, nil
}

func (h *Handler) chatCompletions(w http.ResponseWriter, r *http.Request) {
	_, req, err := h.decodeChatRequest(r)
	if err != nil {
		writeError(w, err)
		return
	}

	if req.Stream {
		h.streamChatCompletions(w, r, req)
		return
	}

	resp, err := h.Chat.Complete(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	out := toSDKChatCompletion(resp)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func toSDKChatCompletion(resp domain.ChatResponse) oai.ChatCompletion {
	choices := make([]oai.ChatCompletionChoice, len(resp.Choices))
	for i, c := range resp.Choices {
		choices[i] = oai.ChatCompletionChoice{
			FinishReason: string(c.FinishReason),
			Index:        int64(c.Index),
			Message: oai.ChatCompletionMessage{
				Content: c.Message.Text(),
				Role:    "assistant",
			},
		}
	}
	return oai.ChatCompletion{
		ID:      resp.ID,
		Choices: choices,
		Created: resp.Created,
		Model:   resp.Model,
		Object:  "chat.completion",
		Usage: oai.CompletionUsage{
			CompletionTokens: int64(resp.Usage.CompletionTokens),
			PromptTokens:     int64(resp.Usage.PromptTokens),
			TotalTokens:      int64(resp.Usage.TotalTokens),
		},
	}
}

// streamChatCompletions writes OpenAI's SSE framing: a series of `data: {...}\n\n` events
// carrying oai.ChatCompletionChunk payloads, terminated by `data: [DONE]\n\n`.
func (h *Handler) streamChatCompletions(w http.ResponseWriter, r *http.Request, req domain.ChatRequest) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, fmt.Errorf("streaming unsupported"))
		return
	}

	ch, err := h.Chat.Stream(r.Context(), req)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	for chunk := range ch {
		sdkChunk := toSDKChatCompletionChunk(chunk)
		data, _ := json.Marshal(sdkChunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}
	fmt.Fprint(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func toSDKChatCompletionChunk(c domain.StreamChunk) oai.ChatCompletionChunk {
	delta := oai.ChatCompletionChunkChoiceDelta{Content: c.DeltaContent}
	if c.DeltaRole != "" {
		delta.Role = string(c.DeltaRole)
	}
	choice := oai.ChatCompletionChunkChoice{
		Delta:        delta,
		FinishReason: string(c.FinishReason),
		Index:        int64(c.Index),
	}
	chunk := oai.ChatCompletionChunk{
		ID:      c.ID,
		Choices: []oai.ChatCompletionChunkChoice{choice},
		Created: c.Created,
		Model:   c.Model,
		Object:  "chat.completion.chunk",
	}
	if c.Usage != nil {
		chunk.Usage = oai.CompletionUsage{
			CompletionTokens: int64(c.Usage.CompletionTokens),
			PromptTokens:     int64(c.Usage.PromptTokens),
			TotalTokens:      int64(c.Usage.TotalTokens),
		}
	}
	return chunk
}
