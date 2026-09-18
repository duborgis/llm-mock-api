// Command openmeter-mock imitates the one HTTP route LiteLLM's OpenMeter integration calls
// (POST /api/v1/events, see litellm/integrations/openmeter.py) and persists every event into
// MongoDB so it can be inspected with any Mongo client, plus a convenience GET endpoint.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	openmeteradapter "github.com/duborgis/llm-mock-api/internal/adapters/inbound/http/openmeter"
	"github.com/duborgis/llm-mock-api/internal/adapters/outbound/mongo"
	"github.com/duborgis/llm-mock-api/internal/openmeter/usecase"
	"github.com/duborgis/llm-mock-api/internal/platform/httpserver"
	"github.com/duborgis/llm-mock-api/internal/platform/logging"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	logger := logging.New()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mongoURI := getenv("OPENMETER_MOCK_MONGO_URI", "mongodb://localhost:27017")
	mongoDB := getenv("OPENMETER_MOCK_MONGO_DB", "openmeter_mock")
	httpPort := getenv("OPENMETER_MOCK_HTTP_PORT", "9010")

	events, err := mongo.NewEventRepo(ctx, mongoURI, mongoDB, "events")
	if err != nil {
		logger.Error("failed to connect to MongoDB", "uri", mongoURI, "err", err)
		os.Exit(1)
	}

	ingestUC := usecase.NewIngest(events)

	mux := http.NewServeMux()
	openmeteradapter.NewRouter(mux, &openmeteradapter.Handler{Events: ingestUC, Logger: logger})
	handler := httpserver.WithAccessLog(logger, "openmeter-mock", mux)

	logger.Info("openmeter-mock started", "http_port", httpPort, "mongo_db", mongoDB)
	if err := httpserver.Run(ctx, ":"+httpPort, handler, logger, "openmeter-mock"); err != nil {
		logger.Error("openmeter-mock server error", "err", err)
	}
	logger.Info("openmeter-mock stopped")
}
