// Package contract holds fidelity tests: they spin up the real mock server and hit it with
// each provider's own official Go SDK client (or, for Vertex, unmarshal into the SDK's own
// structs), proving the wire format is faithful rather than hand-verified.
package contract

import (
	"net/http"
	"net/http/httptest"

	adminhttp "github.com/duborgis/llm-mock-api/internal/adapters/admin/http"
	bedrockadapter "github.com/duborgis/llm-mock-api/internal/adapters/inbound/http/bedrock"
	openaiadapter "github.com/duborgis/llm-mock-api/internal/adapters/inbound/http/openai"
	vertexadapter "github.com/duborgis/llm-mock-api/internal/adapters/inbound/http/vertex"
	"github.com/duborgis/llm-mock-api/internal/adapters/outbound/clock"
	"github.com/duborgis/llm-mock-api/internal/adapters/outbound/scenario"
	"github.com/duborgis/llm-mock-api/internal/core/usecase"
	"github.com/duborgis/llm-mock-api/internal/platform/logging"
)

// newTestServer builds the full hexagonal wiring (same as cmd/server/main.go) and returns
// an httptest.Server plus the scenario store, so tests can seed scenarios directly.
func newTestServer() (*httptest.Server, *scenario.Store) {
	logger := logging.New()
	store := scenario.NewStore()
	models := scenario.NewModelRepo()
	sysClock := clock.New()

	chatUC := usecase.NewChat(store, sysClock)
	embedUC := usecase.NewEmbedding(store, sysClock)
	catalogUC := usecase.NewModelCatalog(models)

	mux := http.NewServeMux()
	openaiadapter.NewRouter(mux, &openaiadapter.Handler{Chat: chatUC, Embed: embedUC, Models: catalogUC, Logger: logger})
	bedrockadapter.NewRouter(mux, &bedrockadapter.Handler{Chat: chatUC, Models: catalogUC, Logger: logger})
	vertexadapter.NewRouter(mux, &vertexadapter.Handler{Chat: chatUC, Models: catalogUC, Logger: logger})
	_ = adminhttp.Handler{} // admin API not needed by contract tests; kept for symmetry with main.go

	return httptest.NewServer(mux), store
}
