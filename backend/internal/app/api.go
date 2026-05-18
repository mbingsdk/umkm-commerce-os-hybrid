package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type API struct {
	deps   *Dependencies
	server *http.Server
}

func NewAPI(deps *Dependencies) *API {
	return &API{
		deps: deps,
		server: &http.Server{
			Addr:              ":" + deps.Config.HTTPPort,
			Handler:           NewRouter(deps),
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (a *API) Run(ctx context.Context) error {
	serverErr := make(chan error, 1)

	go func() {
		a.deps.Logger.Info("api server starting", "addr", a.server.Addr)
		serverErr <- a.server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %w", err)
		}
		return nil
	case <-ctx.Done():
		a.deps.Logger.Info("api server shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown api server: %w", err)
	}

	if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}
