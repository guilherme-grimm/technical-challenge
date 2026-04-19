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

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	"technical-challenge/internal/api"
	"technical-challenge/internal/api/openapi"
	"technical-challenge/internal/config"
	"technical-challenge/internal/domain/gateway"
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

	cfg, err := config.New()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	logger.Info("config loaded",
		zap.String("addr", cfg.Addr),
		zap.String("database", cfg.Database),
		zap.String("collection", cfg.Collection))
	defer func() { _ = logger.Sync() }()

	bootCtx, bootCancel := context.WithTimeout(context.Background(), time.Second*10)
	defer bootCancel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info("connecting to database")
	db, err := setupDatabase(bootCtx, logger, cfg.URI, cfg.Database, cfg.Collection)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() { _ = db.Close(bootCtx) }()

	logger.Info("database connected")

	logger.Info("creating service")
	svc, err := setupService(logger, db)
	if err != nil {
		logger.Fatal("failed to create service", zap.Error(err))
	}

	logger.Info("starting server")
	if err := run(ctx, logger, svc); err != nil {
		logger.Fatal("server exited with error", zap.Error(err))
	}
}

const defaultMaxBytes = 1024 * 1024 * 10

func run(ctx context.Context, logger *zap.Logger, svc gateway.DeviceService) error {
	handler := api.NewHandler(logger, svc)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(api.OpenAPISpec)
	})

	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html charset=utf-8")
		w.Write(api.OpenAPIHTML)
	})
	httpHandler := openapi.HandlerFromMux(handler, mux)

	validator, err := api.OpenAPIValidator(api.OpenAPISpec)
	if err != nil {
		logger.Error("failed to create openapi validator", zap.Error(err))
		return err
	}

	chain := api.Chain(httpHandler,
		api.Recovery(logger),
		api.RequestID(),
		api.RequestLogger(logger),
		api.MaxBytes(defaultMaxBytes),
		api.CORS(),
		validator)
	srv := &http.Server{
		Addr:         defaultAddr,
		Handler:      chain,
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

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed, forcing close", zap.Error(err))
		_ = srv.Close()
		return err
	}

	logger.Info("server stopped")
	return nil
}
