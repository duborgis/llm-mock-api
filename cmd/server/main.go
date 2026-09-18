// Command server wires together the hexagonal mock LLM API: domain usecases behind ports,
// driven by concrete adapters constructed here (the only place that knows about concrete
// adapter types). Two HTTP servers are started: the provider-facing mock API and a small
// admin API for runtime scenario control.
package main

import (
	"context"
	"net/http"
	"os/signal"
	"sync"
	"syscall"

	adminhttp "github.com/duborgis/llm-mock-api/internal/adapters/admin/http"
	bedrockadapter "github.com/duborgis/llm-mock-api/internal/adapters/inbound/http/bedrock"
	openaiadapter "github.com/duborgis/llm-mock-api/internal/adapters/inbound/http/openai"
	vertexadapter "github.com/duborgis/llm-mock-api/internal/adapters/inbound/http/vertex"
	"github.com/duborgis/llm-mock-api/internal/adapters/outbound/clock"
	"github.com/duborgis/llm-mock-api/internal/adapters/outbound/scenario"
	"github.com/duborgis/llm-mock-api/internal/core/usecase"
	"github.com/duborgis/llm-mock-api/internal/platform/config"
	"github.com/duborgis/llm-mock-api/internal/platform/httpserver"
	"github.com/duborgis/llm-mock-api/internal/platform/logging"
)

func main() {
	cfg := config.Load()
	logger := logging.New()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Outbound adapters
	scenarioStore := scenario.NewStore()
	if err := scenario.LoadFile(ctx, cfg.ScenarioFile, scenarioStore); err != nil {
		logger.Error("failed to load scenario file", "path", cfg.ScenarioFile, "err", err)
	}
	modelRepo := scenario.NewModelRepo()
	sysClock := clock.New()

	// Usecases, wired only to ports
	chatUC := usecase.NewChat(scenarioStore, sysClock)
	embedUC := usecase.NewEmbedding(scenarioStore, sysClock)
	imageUC := usecase.NewImage(scenarioStore, sysClock)
	catalogUC := usecase.NewModelCatalog(modelRepo)

	// Inbound HTTP adapters
	mux := http.NewServeMux()
	openaiadapter.NewRouter(mux, &openaiadapter.Handler{Chat: chatUC, Embed: embedUC, Images: imageUC, Models: catalogUC, Logger: logger})
	bedrockadapter.NewRouter(mux, &bedrockadapter.Handler{Chat: chatUC, Models: catalogUC, Logger: logger})
	vertexadapter.NewRouter(mux, &vertexadapter.Handler{Chat: chatUC, Models: catalogUC, Logger: logger})

	adminMux := http.NewServeMux()
	adminhttp.NewRouter(adminMux, &adminhttp.Handler{Scenarios: scenarioStore, Logger: logger})

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		handler := httpserver.WithAccessLog(logger, "mock-api", mux)
		if err := httpserver.Run(ctx, ":"+cfg.HTTPPort, handler, logger, "mock-api"); err != nil {
			logger.Error("mock-api server error", "err", err)
		}
	}()
	go func() {
		defer wg.Done()
		handler := httpserver.WithAccessLog(logger, "admin-api", adminMux)
		if err := httpserver.Run(ctx, ":"+cfg.AdminPort, handler, logger, "admin-api"); err != nil {
			logger.Error("admin-api server error", "err", err)
		}
	}()

	logger.Info("llm-mock-api started", "http_port", cfg.HTTPPort, "admin_port", cfg.AdminPort)
	<-ctx.Done()
	wg.Wait()
	logger.Info("llm-mock-api stopped")
}
