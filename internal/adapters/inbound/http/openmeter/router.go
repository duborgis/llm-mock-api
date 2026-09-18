package openmeter

import (
	"log/slog"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/openmeter/ports"
)

type Handler struct {
	Events ports.EventIngestUseCase
	Logger *slog.Logger
}

// NewRouter registers the routes LiteLLM's OpenMeter integration and a human actually need:
//   - POST /api/v1/events — what litellm/integrations/openmeter.py posts on every successful
//     call (see OPENMETER_API_ENDPOINT). Real OpenMeter expects "application/cloudevents+json";
//     this mock accepts plain JSON too, for convenience.
//   - GET  /api/v1/events — not part of OpenMeter's real API; a convenience endpoint to
//     inspect the last N ingested events without needing a Mongo client.
func NewRouter(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /api/v1/events", h.ingestEvent)
	mux.HandleFunc("GET /api/v1/events", h.listEvents)
}
