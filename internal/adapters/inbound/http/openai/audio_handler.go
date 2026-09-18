package openai

import (
	"encoding/json"
	"net/http"

	oai "github.com/openai/openai-go"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// createTranscription decodes a multipart/form-data request (the wire format
// oai.AudioTranscriptionNewParams.MarshalMultipart produces) and encodes the response as a
// real oai.Transcription, so fidelity is structural just like the JSON-bodied routes.
func (h *Handler) createTranscription(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		writeError(w, h.Logger, domain.ErrInvalidRequest)
		return
	}

	model := r.FormValue("model")
	if model == "" {
		writeError(w, h.Logger, domain.ErrInvalidRequest)
		return
	}
	_, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, h.Logger, domain.ErrInvalidRequest)
		return
	}

	resp, err := h.Transcriptions.Transcribe(r.Context(), domain.TranscriptionRequest{
		Model:    model,
		Filename: header.Filename,
		Language: r.FormValue("language"),
		Prompt:   r.FormValue("prompt"),
	})
	if err != nil {
		writeError(w, h.Logger, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(oai.Transcription{Text: resp.Text})
}

// createSpeech decodes a real oai.AudioSpeechNewParams and streams back raw audio bytes, since
// that's what AudioSpeechService.New itself returns (*http.Response, not a JSON struct).
func (h *Handler) createSpeech(w http.ResponseWriter, r *http.Request) {
	var params oai.AudioSpeechNewParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		writeError(w, h.Logger, domain.ErrInvalidRequest)
		return
	}
	if params.Input == "" || params.Model == "" {
		writeError(w, h.Logger, domain.ErrInvalidRequest)
		return
	}

	resp, err := h.Speech.Synthesize(r.Context(), domain.SpeechRequest{
		Model:          params.Model,
		Input:          params.Input,
		Voice:          string(params.Voice),
		ResponseFormat: string(params.ResponseFormat),
	})
	if err != nil {
		writeError(w, h.Logger, err)
		return
	}

	w.Header().Set("Content-Type", resp.ContentType)
	_, _ = w.Write(resp.Audio)
}
