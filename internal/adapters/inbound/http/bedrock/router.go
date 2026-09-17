// Package bedrock is the driving HTTP adapter that speaks AWS Bedrock Runtime's wire contract
// (Converse/ConverseStream, InvokeModel/InvokeModelWithResponseStream for Anthropic Claude bodies)
// plus a couple of Bedrock control-plane endpoints for model listing.
//
// Note on fidelity strategy: the bedrockruntime Go SDK's request/response Go structs
// (ConverseInput/Output, InvokeModelInput/Output, etc.) have NO encoding/json tags — the real
// SDK serializes/deserializes them through AWS's smithy restjson1 protocol machinery, not
// encoding/json. Reusing those structs directly for our own JSON marshal/unmarshal is therefore
// not viable. Instead we hand-roll small JSON-tagged mirror structs that match Bedrock's
// documented wire shape field-for-field (messages[].role/content[].text, output.message,
// usage.inputTokens/outputTokens, etc). Fidelity is validated in test/contract/bedrock_contract_test.go
// by pointing the REAL bedrockruntime.Client at this server and confirming its own smithy
// deserializer parses our JSON into ConverseOutput/InvokeModelOutput without error.
package bedrock

import (
	"log/slog"
	"net/http"

	"github.com/duborgis/llm-mock-api/internal/core/ports"
)

type Handler struct {
	Chat   ports.ChatCompletionUseCase
	Models ports.ModelCatalogUseCase
	Logger *slog.Logger
}

func NewRouter(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /model/{modelId}/invoke", h.invokeModel)
	mux.HandleFunc("POST /model/{modelId}/invoke-with-response-stream", h.invokeModelStream)
	mux.HandleFunc("POST /model/{modelId}/converse", h.converse)
	mux.HandleFunc("POST /model/{modelId}/converse-stream", h.converseStream)
	mux.HandleFunc("GET /foundation-models", h.listFoundationModels)
	mux.HandleFunc("GET /foundation-models/{modelId}", h.getFoundationModel)
}
