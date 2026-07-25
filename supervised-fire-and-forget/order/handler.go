// Package order serves order requests and queues their background work.
package order

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/rednafi/examples/supervised-fire-and-forget/internal/bg"
)

type requestIDKey struct{}

// Tasks contains the background operations started by the order handler.
type Tasks interface {
	SendNotification(context.Context, string) error
	WriteDiagnosticLog(context.Context, string) error
}

// Operations provides the example background operations.
type Operations struct{}

// SendNotification sends an order notification.
func (Operations) SendNotification(ctx context.Context, user string) error {
	return wait(ctx, 10*time.Millisecond)
}

// WriteDiagnosticLog writes the order diagnostic log.
func (Operations) WriteDiagnosticLog(ctx context.Context, user string) error {
	return wait(ctx, 20*time.Millisecond)
}

type handler struct {
	background  *bg.Background
	tasks       Tasks
	logger      *slog.Logger
	taskTimeout time.Duration
}

// NewHandler returns an HTTP handler that queues best-effort background work.
func NewHandler(
	background *bg.Background,
	tasks Tasks,
	logger *slog.Logger,
	taskTimeout time.Duration,
) (http.Handler, error) {
	switch {
	case background == nil:
		return nil, errors.New("background is nil")
	case tasks == nil:
		return nil, errors.New("order tasks are nil")
	case logger == nil:
		return nil, errors.New("logger is nil")
	case taskTimeout <= 0:
		return nil, fmt.Errorf("task timeout must be positive: %s", taskTimeout)
	}

	h := &handler{
		background:  background,
		tasks:       tasks,
		logger:      logger,
		taskTimeout: taskTimeout,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", h.createOrder)
	return mux, nil
}

func (h *handler) createOrder(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		http.Error(w, "missing user", http.StatusBadRequest)
		return
	}

	requestCtx := r.Context()
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		requestCtx = context.WithValue(requestCtx, requestIDKey{}, requestID)
	}
	detached := context.WithoutCancel(requestCtx)

	if !h.background.Submit(func() {
		ctx, cancel := context.WithTimeout(context.Background(), h.taskTimeout)
		defer cancel()

		if err := h.tasks.SendNotification(ctx, user); err != nil {
			h.logger.ErrorContext(ctx, "send notification", "error", err)
			return
		}
		h.logger.InfoContext(ctx, "notification sent", "user", user)
	}) {
		h.logger.WarnContext(requestCtx, "notification task rejected", "user", user)
	}

	if !h.background.Submit(func() {
		ctx, cancel := context.WithTimeout(detached, h.taskTimeout)
		defer cancel()

		if err := h.tasks.WriteDiagnosticLog(ctx, user); err != nil {
			h.logger.ErrorContext(
				ctx,
				"write diagnostic log",
				"request_id", requestIDFromContext(ctx),
				"error", err,
			)
			return
		}
		h.logger.InfoContext(
			ctx,
			"diagnostic log written",
			"user", user,
			"request_id", requestIDFromContext(ctx),
		)
	}) {
		h.logger.WarnContext(
			requestCtx,
			"diagnostic log task rejected",
			"user", user,
			"request_id", requestIDFromContext(requestCtx),
		)
	}

	w.WriteHeader(http.StatusAccepted)
}

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}
