package openai

import (
	"encoding/json"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// videoCreateRequest mirrors openai-python's VideoCreateParams (openai.types.video_create_params)
// — the Go SDK has no video support to decode against, so this struct is hand-kept in sync with
// the Python one instead. Fields not needed for a mock (input_reference) are omitted.
type videoCreateRequest struct {
	Prompt  string `json:"prompt"`
	Model   string `json:"model"`
	Seconds string `json:"seconds"`
	Size    string `json:"size"`
}

// videoResponse mirrors openai-python's Video type (openai.types.video).
type videoResponse struct {
	ID          string          `json:"id"`
	Object      string          `json:"object"`
	Status      string          `json:"status"`
	Progress    int             `json:"progress"`
	Model       string          `json:"model"`
	Prompt      string          `json:"prompt,omitempty"`
	Seconds     string          `json:"seconds"`
	Size        string          `json:"size"`
	CreatedAt   int64           `json:"created_at"`
	CompletedAt *int64          `json:"completed_at,omitempty"`
	ExpiresAt   *int64          `json:"expires_at,omitempty"`
	Error       *videoErrorBody `json:"error,omitempty"`
}

type videoErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func toVideoResponse(v domain.Video) videoResponse {
	resp := videoResponse{
		ID:          v.ID,
		Object:      v.Object,
		Status:      v.Status,
		Progress:    v.Progress,
		Model:       v.Model,
		Prompt:      v.Prompt,
		Seconds:     v.Seconds,
		Size:        v.Size,
		CreatedAt:   v.CreatedAt,
		CompletedAt: v.CompletedAt,
		ExpiresAt:   v.ExpiresAt,
	}
	if v.Error != nil {
		resp.Error = &videoErrorBody{Code: v.Error.Code, Message: v.Error.Message}
	}
	return resp
}

func (h *Handler) createVideo(w http.ResponseWriter, r *http.Request) {
	var req videoCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, h.Logger, domain.ErrInvalidRequest)
		return
	}

	job, err := h.Videos.Create(r.Context(), domain.VideoRequest{
		Model:   req.Model,
		Prompt:  req.Prompt,
		Seconds: req.Seconds,
		Size:    req.Size,
	})
	if err != nil {
		writeError(w, h.Logger, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toVideoResponse(job))
}

func (h *Handler) getVideo(w http.ResponseWriter, r *http.Request) {
	job, err := h.Videos.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, h.Logger, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toVideoResponse(job))
}

func (h *Handler) getVideoContent(w http.ResponseWriter, r *http.Request) {
	content, contentType, err := h.Videos.Content(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, h.Logger, err)
		return
	}

	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(content)
}
