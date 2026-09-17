package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Run starts an http.Server on addr with handler, blocks until ctx is cancelled, then
// gracefully shuts it down. It is a generic bootstrap with no provider-specific knowledge.
func Run(ctx context.Context, addr string, handler http.Handler, logger *slog.Logger, name string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting server", "name", name, "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		logger.Info("shutting down server", "name", name)
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
