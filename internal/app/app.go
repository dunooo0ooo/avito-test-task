package app

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/dunooo0ooo/avito-test-task/pkg/config"
	"go.uber.org/zap"
)

type App struct {
	srv    *http.Server
	logger *zap.Logger
	cfg    *config.Config
}

func NewApp(srv *http.Server, logger *zap.Logger, cfg *config.Config) *App {
	return &App{
		srv:    srv,
		logger: logger,
		cfg:    cfg,
	}
}

func (a *App) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		a.logger.Info("starting HTTP server",
			zap.String("addr", a.srv.Addr),
		)

		if err := a.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

func (a *App) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	a.logger.Info("shutting down HTTP server")

	if err := a.srv.Shutdown(ctx); err != nil {
		a.logger.Error("failed to shutdown HTTP server", zap.Error(err))
		return err
	}

	a.logger.Info("HTTP server stopped")
	return nil
}
