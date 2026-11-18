package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/panjitamzil/scalens/internal/observability"
	"github.com/panjitamzil/scalens/internal/server"
	"github.com/panjitamzil/scalens/services/payment"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("payment stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	delay, err := server.IntEnv("PAYMENT_DELAY_MS", 5, 0, 10000)
	if err != nil {
		return err
	}
	shutdown, err := observability.Tracing(context.Background(), "payment-service")
	if err != nil {
		return err
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			slog.Warn("trace shutdown", "error", err)
		}
	}()
	return server.Run(context.Background(), payment.Handler(observability.NewMetrics("payment-service"), time.Duration(delay)*time.Millisecond, os.Getenv("PAYMENT_FAIL") == "true"), server.Env("HTTP_ADDR", ":8081"))
}
