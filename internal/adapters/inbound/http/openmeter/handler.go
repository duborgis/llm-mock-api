package openmeter

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/duborgis/llm-mock-api/internal/openmeter/domain"
)

func (h *Handler) ingestEvent(w http.ResponseWriter, r *http.Request) {
	var e domain.Event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		if h.Logger != nil {
			h.Logger.Error("openmeter: failed to decode event", "err", err)
		}
		http.Error(w, `{"error":"invalid event body"}`, http.StatusBadRequest)
		return
	}

	if err := h.Events.Ingest(r.Context(), e); err != nil {
		if h.Logger != nil {
			h.Logger.Error("openmeter: failed to store event", "err", err)
		}
		http.Error(w, `{"error":"failed to store event"}`, http.StatusInternalServerError)
		return
	}

	if h.Logger != nil {
		h.Logger.Info("openmeter: event ingested", "id", e.ID, "type", e.Type, "subject", e.Subject, "model", e.Data["model"])
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) listEvents(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}

	events, err := h.Events.List(r.Context(), limit)
	if err != nil {
		if h.Logger != nil {
			h.Logger.Error("openmeter: failed to list events", "err", err)
		}
		http.Error(w, `{"error":"failed to list events"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Events []domain.Event `json:"events"`
	}{Events: events})
}
