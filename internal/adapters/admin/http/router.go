// Package http (admin) exposes a small non-provider API for configuring mock behavior at
// runtime: canned responses, latency injection, and error injection, keyed by model+operation.
// This is purely a test-control surface, not part of any provider's wire contract.
package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/domain"
	"github.com/duborgis/llm-mock-api/internal/core/ports"
)

type Handler struct {
	Scenarios ports.ScenarioRepository
	Logger    *slog.Logger
}

func NewRouter(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /admin/scenarios", h.list)
	mux.HandleFunc("PUT /admin/scenarios", h.put)
}

// scenarioDTO is the admin API's wire shape: plain, provider-agnostic JSON.
type scenarioDTO struct {
	Model           string `json:"model"`
	Operation       string `json:"operation"`
	ResponseText    string `json:"response_text,omitempty"`
	LatencyMs       int    `json:"latency_ms,omitempty"`
	TokensPerSecond int    `json:"tokens_per_second,omitempty"`
	Error           *struct {
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
		Type       string `json:"type"`
	} `json:"error,omitempty"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	scenarios, err := h.Scenarios.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(scenarios)
}

func (h *Handler) put(w http.ResponseWriter, r *http.Request) {
	var dto scenarioDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "invalid scenario body", http.StatusBadRequest)
		return
	}
	if dto.Model == "" {
		http.Error(w, "model is required", http.StatusBadRequest)
		return
	}
	op := dto.Operation
	if op == "" {
		op = "chat"
	}

	sc := domain.Scenario{
		Key:             dto.Model + "|" + op,
		ResponseText:    dto.ResponseText,
		LatencyMs:       dto.LatencyMs,
		TokensPerSecond: dto.TokensPerSecond,
	}
	if dto.Error != nil {
		sc.Error = &domain.ErrorInjection{
			StatusCode: dto.Error.StatusCode,
			Message:    dto.Error.Message,
			Type:       dto.Error.Type,
		}
	}

	if err := h.Scenarios.Put(r.Context(), sc); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
