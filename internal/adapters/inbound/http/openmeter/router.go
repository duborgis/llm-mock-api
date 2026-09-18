package openmeter

import (
	"log/slog"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/openmeter/ports"
)

type Handler struct {
	Events    ports.EventIngestUseCase
	Responses ports.RawResponseIngestUseCase
	Logger    *slog.Logger
}

// NewRouter registers the routes LiteLLM's OpenMeter integration and a human actually need:
//   - POST /api/v1/events         — what litellm/integrations/openmeter.py posts on every
//     successful call (see OPENMETER_API_ENDPOINT). Real OpenMeter expects
//     "application/cloudevents+json"; this mock accepts plain JSON too, for convenience.
//   - GET  /api/v1/events         — not part of OpenMeter's real API; a convenience endpoint
//     to inspect the last N ingested events without needing a Mongo client.
//   - POST /api/v1/raw-responses — not part of OpenMeter either; posted by our own custom
//     LiteLLM callback (litellm/custom_callback.py), which forwards kwargs["original_response"]
//     for every call so the full provider-shaped response can be inspected later.
//   - GET  /api/v1/raw-responses — convenience endpoint mirroring GET /api/v1/events.
func NewRouter(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /api/v1/events", h.ingestEvent)
	mux.HandleFunc("GET /api/v1/events", h.listEvents)
	mux.HandleFunc("POST /api/v1/raw-responses", h.ingestRawResponse)
	mux.HandleFunc("GET /api/v1/raw-responses", h.listRawResponses)
}
