package openai

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
)

// openAIErrorBody matches OpenAI's documented error envelope shape.
type openAIErrorBody struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param,omitempty"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}

func writeError(w http.ResponseWriter, logger *slog.Logger, err error) {
	status := http.StatusInternalServerError
	body := openAIErrorBody{}
	body.Error.Message = err.Error()
	body.Error.Type = "internal_error"

	var injected *domain.InjectedError
	switch {
	case errors.As(err, &injected):
		status = injected.Injection.StatusCode
		if status == 0 {
			status = http.StatusInternalServerError
		}
		body.Error.Type = injected.Injection.Type
		body.Error.Message = injected.Injection.Message
	case errors.Is(err, domain.ErrInvalidRequest):
		status = http.StatusBadRequest
		body.Error.Type = "invalid_request_error"
	case errors.Is(err, domain.ErrModelNotFound):
		status = http.StatusNotFound
		body.Error.Type = "invalid_request_error"
	}

	if logger != nil {
		logger.Error("openai: request failed", "status", status, "type", body.Error.Type, "message", body.Error.Message)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
