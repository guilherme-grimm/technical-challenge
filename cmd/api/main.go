// Command api is the entrypoint for the Devices API. It wires the logger,
// HTTP handler, and server, then blocks until SIGINT/SIGTERM triggers a
// graceful shutdown. Config, Mongo, middleware, and validation are not
// wired yet — each TODO below marks a seam to be filled in.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"technical-challenge/internal/api"
	"technical-challenge/internal/api/openapi"
)

// Defaults to avoid badly populated config
const (
	defaultAddr            = ":8080"
	defaultShutdownTimeout = 15 * time.Second
	defaultReadTimeout     = 15 * time.Second
	defaultWriteTimeout    = 15 * time.Second
	defaultIdleTimeout     = 60 * time.Second
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	if err := run(logger); err != nil {
		logger.Fatal("server exited with error", zap.Error(err))
	}
}

func run(logger *zap.Logger) error {
	// TODO: connect to Mongo, ping, defer client.Disconnect. injecting svcs here
	handler := api.NewHandler(logger)

	mux := http.NewServeMux()

	// TODO: register GET /docs (Swagger UI) and GET /openapi.yaml before wiring the generated routes
	httpHandler := openapi.HandlerFromMux(handler, mux)

	// TODO: wrap httpHandler with the middleware chain: recovery -> request ID -> logging -> MaxBytesReader -> CORS -> kin-openapi validator -> routing
	srv := &http.Server{
		Addr:         defaultAddr,
		Handler:      httpHandler,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
		IdleTimeout:  defaultIdleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting http server", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))
	case err := <-serverErr:
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed, forcing close", zap.Error(err))
		_ = srv.Close()
		return err
	}

	logger.Info("server stopped")
	return nil
}
