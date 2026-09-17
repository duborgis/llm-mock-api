package openai

import (
	"encoding/json"
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

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	body := openAIErrorBody{}
	body.Error.Message = err.Error()
	body.Error.Type = "internal_error"

	switch e := err.(type) {
	case *domain.InjectedError:
		status = e.Injection.StatusCode
		if status == 0 {
			status = http.StatusInternalServerError
		}
		body.Error.Type = e.Injection.Type
		body.Error.Message = e.Injection.Message
	default:
		if err == domain.ErrInvalidRequest {
			status = http.StatusBadRequest
			body.Error.Type = "invalid_request_error"
		} else if err == domain.ErrModelNotFound {
			status = http.StatusNotFound
			body.Error.Type = "invalid_request_error"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
