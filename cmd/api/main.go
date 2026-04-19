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
	"technical-challenge/internal/resource/database"
)

const defaultMaxBytes = 1024 * 1024 * 10

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = logger.Sync() }()

	cfg, err := config.New()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	logger.Info("config loaded",
		zap.String("addr", cfg.Addr),
		zap.String("database", cfg.Database),
		zap.String("collection", cfg.Collection))

	bootCtx, bootCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer bootCancel()

	logger.Info("connecting to database")
	db, err := setupDatabase(bootCtx, logger, cfg.URI, cfg.Database, cfg.Collection)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer closeCancel()
		_ = db.Close(closeCtx)
	}()

	logger.Info("database connected")

	svc, err := setupService(logger, db)
	if err != nil {
		logger.Fatal("failed to create service", zap.Error(err))
	}

	if err := run(logger, cfg, svc, db); err != nil {
		logger.Fatal("server exited with error", zap.Error(err))
	}
}

func run(logger *zap.Logger, cfg *config.Config, svc gateway.DeviceService, db database.Service) error {
	handler := api.NewHandler(logger, svc, db)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(api.OpenAPISpec)
	})

	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(api.OpenAPIHTML)
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
		Addr:         cfg.Addr,
		Handler:      chain,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed, forcing close", zap.Error(err))
		_ = srv.Close()
		return err
	}

	logger.Info("server stopped")
	return nil
}
