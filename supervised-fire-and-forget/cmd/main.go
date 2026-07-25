package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/rednafi/examples/supervised-fire-and-forget/internal/bg"
	"github.com/rednafi/examples/supervised-fire-and-forget/order"
)

const (
	backgroundCapacity = 64
	backgroundWorkers  = 4
	taskTimeout        = 250 * time.Millisecond
	shutdownTimeout    = 2 * time.Second
)

type runConfig struct {
	address string
	logger  *slog.Logger
	ready   chan<- string
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(context.Background(), runConfig{
		address: "127.0.0.1:8080",
		logger:  logger,
	}); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(parent context.Context, cfg runConfig) error {
	switch {
	case parent == nil:
		return errors.New("parent context is nil")
	case cfg.address == "":
		return errors.New("address is empty")
	case cfg.logger == nil:
		return errors.New("logger is nil")
	}

	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer stop()

	listener, err := net.Listen("tcp", cfg.address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", cfg.address, err)
	}
	defer listener.Close()

	logPanic := func(value any) {
		cfg.logger.Error(
			"background task panicked",
			"panic", value,
			"stack", string(debug.Stack()),
		)
	}
	background := bg.New(backgroundCapacity, backgroundWorkers, logPanic)
	defer background.Stop()

	handler, err := order.NewHandler(
		background,
		order.Operations{},
		cfg.logger,
		taskTimeout,
	)
	if err != nil {
		return fmt.Errorf("create HTTP handler: %w", err)
	}
	server := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: time.Second,
	}
	// Drains active handlers even when Serve fails before a signal arrives, so their
	// Submit calls finish before the deferred background.Stop. A second Shutdown
	// after the explicit one below is a no-op.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.WithoutCancel(ctx),
			shutdownTimeout,
		)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()

	address := listener.Addr().String()
	cfg.logger.Info("server listening", "address", address)
	if cfg.ready != nil {
		select {
		case cfg.ready <- address:
		case <-ctx.Done():
		}
	}

	select {
	case err := <-serveErr:
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return fmt.Errorf("shut down HTTP server: %w", err)
	}

	if err := <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP during shutdown: %w", err)
	}
	return nil
}
