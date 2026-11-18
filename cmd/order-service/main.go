package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panjitamzil/scalens/internal/observability"
	"github.com/panjitamzil/scalens/internal/server"
	"github.com/panjitamzil/scalens/services/order"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	if err := run(); err != nil {
		slog.Error("order stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	work, err := server.IntEnv("ORDER_SYNTHETIC_WORK", 0, 0, 100000000)
	if err != nil {
		return err
	}
	paymentTimeout, err := server.IntEnv("PAYMENT_TIMEOUT_MS", 2000, 10, 30000)
	if err != nil {
		return err
	}
	maxConns, err := server.IntEnv("DB_POOL_MAX", 8, 1, 64)
	if err != nil {
		return err
	}
	cfg, err := pgxpool.ParseConfig(server.Env("DATABASE_URL", "postgres://scalens:scalens@localhost:5432/scalens?sslmode=disable"))
	if err != nil {
		return err
	}
	cfg.MaxConns = int32(maxConns)
	cfg.ConnConfig.ConnectTimeout = 3 * time.Second
	cfg.ConnConfig.RuntimeParams["statement_timeout"] = "5000"
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for {
		if err := pool.Ping(ctx); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	shutdown, err := observability.Tracing(context.Background(), "order-service")
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
	s := order.New(pool, observability.NewMetrics("order-service"), server.Env("PAYMENT_URL", "http://localhost:8081"), work)
	s.Client.Timeout = time.Duration(paymentTimeout) * time.Millisecond
	return server.Run(context.Background(), s.Handler(), server.Env("HTTP_ADDR", ":8080"))
}
