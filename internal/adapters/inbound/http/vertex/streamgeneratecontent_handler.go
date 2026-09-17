package vertex

import (
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/genai"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// streamGenerateContent implements Vertex's streaming contract: a JSON ARRAY of
// GenerateContentResponse objects sent incrementally over chunked transfer encoding
// (NOT Server-Sent Events / `data:` framing, unlike OpenAI). We open the array, write one
// element per domain.StreamChunk, comma-separating, then close it.
func (h *Handler) streamGenerateContent(w http.ResponseWriter, r *http.Request, model string) {
	var body generateContentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.Contents) == 0 {
		writeVertexError(w, domain.ErrInvalidRequest)
		return
	}
	req := toDomainRequest(model, body.Contents, body.SystemInstruction)

	ch, err := h.Chat.Stream(r.Context(), req)
	if err != nil {
		writeVertexError(w, err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeVertexError(w, fmt.Errorf("streaming unsupported"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "[")
	flusher.Flush()

	first := true
	for chunk := range ch {
		resp := &genai.GenerateContentResponse{
			ModelVersion: chunk.Model,
			ResponseID:   chunk.ID,
			Candidates: []*genai.Candidate{{
				Content:      &genai.Content{Role: "model", Parts: []*genai.Part{{Text: chunk.DeltaContent}}},
				FinishReason: finishReasonFor(chunk.FinishReason),
				Index:        int32(chunk.Index),
			}},
		}
		if chunk.Usage != nil {
			resp.UsageMetadata = &genai.GenerateContentResponseUsageMetadata{
				PromptTokenCount:     int32(chunk.Usage.PromptTokens),
				CandidatesTokenCount: int32(chunk.Usage.CompletionTokens),
				TotalTokenCount:      int32(chunk.Usage.TotalTokens),
			}
		}
		data, _ := json.Marshal(resp)
		if !first {
			fmt.Fprint(w, ",")
		}
		first = false
		w.Write(data)
		flusher.Flush()
	}

	fmt.Fprint(w, "]")
	flusher.Flush()
}
