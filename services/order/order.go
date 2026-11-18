package order

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panjitamzil/scalens/internal/observability"
	"github.com/panjitamzil/scalens/internal/server"
	"github.com/panjitamzil/scalens/services/payment"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Item struct {
	Price    int64 `json:"price"`
	Quantity int64 `json:"quantity"`
}

type Request struct {
	CustomerID int64  `json:"customer_id"`
	Items      []Item `json:"items"`
}

type Order struct {
	ID         string `json:"id"`
	CustomerID int64  `json:"customer_id"`
	Amount     int64  `json:"amount"`
	Recent     int64  `json:"recent_orders"`
	Status     string `json:"status"`
}

type Service struct {
	Pool       *pgxpool.Pool
	Metrics    *observability.Metrics
	PaymentURL string
	Client     *http.Client
	Work       int
	ready      atomic.Bool
}

func Total(req Request) (int64, error) {
	if req.CustomerID <= 0 || req.CustomerID > 10000 || len(req.Items) == 0 || len(req.Items) > 100 {
		return 0, errors.New("invalid customer or items")
	}
	var subtotal int64
	for _, item := range req.Items {
		if item.Price <= 0 || item.Price > 100000000 || item.Quantity <= 0 || item.Quantity > 1000 {
			return 0, errors.New("invalid item price or quantity")
		}
		subtotal += item.Price * item.Quantity
	}
	if subtotal >= 10000 {
		subtotal -= subtotal / 20
	}
	return subtotal + subtotal/10, nil
}

func New(pool *pgxpool.Pool, m *observability.Metrics, paymentURL string, work int) *Service {
	s := &Service{Pool: pool, Metrics: m, PaymentURL: paymentURL, Work: work, Client: &http.Client{Timeout: 2 * time.Second, Transport: otelhttp.NewTransport(http.DefaultTransport)}}
	for _, state := range []string{"idle", "in_use"} {
		m.Registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "db_pool_connections", Help: "Pool connections.", ConstLabels: prometheus.Labels{"service": "order-service", "state": state}}, func() float64 {
			if state == "idle" {
				return float64(pool.Stat().IdleConns())
			}
			return float64(pool.Stat().AcquiredConns())
		}))
	}
	m.Registry.MustRegister(
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{Name: "db_pool_max_connections", Help: "Pool capacity.", ConstLabels: prometheus.Labels{"service": "order-service"}}, func() float64 { return float64(pool.Stat().MaxConns()) }),
		prometheus.NewCounterFunc(prometheus.CounterOpts{Name: "db_pool_empty_acquires_total", Help: "Acquires waiting for a connection.", ConstLabels: prometheus.Labels{"service": "order-service"}}, func() float64 { return float64(pool.Stat().EmptyAcquireCount()) }),
		prometheus.NewCounterFunc(prometheus.CounterOpts{Name: "db_pool_acquire_wait_seconds_total", Help: "Cumulative wait for an empty pool.", ConstLabels: prometheus.Labels{"service": "order-service"}}, func() float64 { return pool.Stat().EmptyAcquireWaitTime().Seconds() }),
	)
	return s
}

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /orders", s.Metrics.Wrap("/orders", s.create))
	mux.Handle("GET /orders/{id}", s.Metrics.Wrap("/orders/{id}", s.get))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		// A busy pool must not take a started pod out of rotation.
		if !s.ready.Load() {
			ctx, cancel := context.WithTimeout(r.Context(), time.Second)
			defer cancel()
			if err := s.Pool.Ping(ctx); err != nil {
				w.WriteHeader(503)
				return
			}
			s.ready.Store(true)
		}
		w.WriteHeader(200)
	})
	mux.Handle("GET /metrics", s.Metrics.Handler())
	return mux
}

func (s *Service) operation(ctx context.Context, name string, fn func(context.Context, *pgxpool.Conn) error) error {
	acquireCtx, acquire := otel.Tracer("scalens/order").Start(ctx, "Database.Acquire", trace.WithAttributes(attribute.String("dbys.phase", "connection_acquire")))
	start := time.Now()
	conn, err := s.Pool.Acquire(acquireCtx)
	s.Metrics.Acquire.WithLabelValues(outcome(err)).Observe(time.Since(start).Seconds())
	if err != nil {
		acquire.RecordError(err)
		acquire.SetStatus(codes.Error, "connection acquisition failed")
	}
	acquire.End()
	if err != nil {
		return err
	}
	defer conn.Release()
	ctx, query := otel.Tracer("scalens/order").Start(ctx, name, trace.WithAttributes(attribute.String("dbys.phase", "query_execute"), attribute.String("db.system.name", "postgresql"), attribute.String("db.query.summary", name)))
	start = time.Now()
	err = fn(ctx, conn)
	s.Metrics.Query.WithLabelValues(name, outcome(err)).Observe(time.Since(start).Seconds())
	if err != nil {
		query.RecordError(err)
		query.SetStatus(codes.Error, "database operation failed")
	}
	query.End()
	return err
}

func outcome(err error) string {
	if err != nil {
		return "failure"
	}
	return "success"
}

func (s *Service) create(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	ctx, span := otel.Tracer("scalens/order").Start(ctx, "OrderService.CreateOrder")
	defer span.End()
	var req Request
	if err := server.Decode(w, r, &req); err != nil {
		server.JSON(w, 400, map[string]string{"error": "invalid JSON"})
		return
	}
	amount, err := Total(req)
	if err != nil {
		server.JSON(w, 400, map[string]string{"error": err.Error()})
		return
	}
	var id [16]byte
	if _, err = rand.Read(id[:]); err != nil {
		server.JSON(w, 500, map[string]string{"error": "cannot create order"})
		return
	}
	if s.Work > 0 {
		workCtx, workSpan := otel.Tracer("scalens/order").Start(ctx, "OrderService.SyntheticWork", trace.WithAttributes(attribute.String("dbys.phase", "synthetic_work")))
		digest := sha256.Sum256(id[:])
		for i := 0; i < s.Work; i++ {
			if i%1024 == 0 && workCtx.Err() != nil {
				err = workCtx.Err()
				break
			}
			digest = sha256.Sum256(digest[:])
		}
		workSpan.SetAttributes(attribute.Int("work.checksum", int(digest[0])))
		workSpan.End()
		if err != nil {
			server.JSON(w, 504, map[string]string{"error": "order deadline"})
			return
		}
	}
	o := Order{ID: hex.EncodeToString(id[:]), CustomerID: req.CustomerID, Amount: amount, Status: "pending"}
	err = s.operation(ctx, "OrderRepository.CountRecentOrders", func(ctx context.Context, c *pgxpool.Conn) error {
		return c.QueryRow(ctx, "SELECT count(*) FROM orders WHERE customer_id=$1 AND created_at >= TIMESTAMPTZ '2026-01-01' AND created_at < TIMESTAMPTZ '2026-02-01'", req.CustomerID).Scan(&o.Recent)
	})
	if err == nil {
		err = s.operation(ctx, "OrderRepository.SaveOrder", func(ctx context.Context, c *pgxpool.Conn) error {
			_, err := c.Exec(ctx, "INSERT INTO orders(id,customer_id,amount,recent_orders,status) VALUES($1,$2,$3,$4,$5)", o.ID, o.CustomerID, o.Amount, o.Recent, o.Status)
			return err
		})
	}
	if err != nil {
		span.RecordError(err)
		server.JSON(w, 503, map[string]string{"error": "database unavailable"})
		return
	}
	payload, _ := json.Marshal(payment.Request{OrderID: o.ID, Amount: amount})
	callCtx, callSpan := otel.Tracer("scalens/order").Start(ctx, "PaymentClient.Charge")
	payReq, err := http.NewRequestWithContext(callCtx, http.MethodPost, s.PaymentURL+"/payments", bytes.NewReader(payload))
	if err != nil {
		callSpan.End()
		server.JSON(w, 500, map[string]string{"error": "invalid payment endpoint"})
		return
	}
	payReq.Header.Set("Content-Type", "application/json")
	resp, callErr := s.Client.Do(payReq)
	status := 201
	o.Status = "paid"
	if callErr != nil {
		o.Status, status = "payment_unknown", 504
		callSpan.RecordError(callErr)
	} else {
		var result struct {
			Status string `json:"status"`
		}
		decodeErr := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&result)
		resp.Body.Close()
		if resp.StatusCode != 200 || decodeErr != nil || result.Status != "paid" {
			o.Status, status = "failed", 502
		}
	}
	callSpan.End()
	err = s.operation(ctx, "OrderRepository.UpdateStatus", func(ctx context.Context, c *pgxpool.Conn) error {
		_, err := c.Exec(ctx, "UPDATE orders SET status=$2 WHERE id=$1", o.ID, o.Status)
		return err
	})
	if err != nil {
		span.RecordError(err)
		server.JSON(w, 503, map[string]string{"error": "order status not persisted", "id": o.ID})
		return
	}
	server.JSON(w, status, o)
}

func (s *Service) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if len(id) != 32 {
		server.JSON(w, 400, map[string]string{"error": "invalid order id"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	var o Order
	err := s.operation(ctx, "OrderRepository.GetOrder", func(ctx context.Context, c *pgxpool.Conn) error {
		return c.QueryRow(ctx, "SELECT id,customer_id,amount,recent_orders,status FROM orders WHERE id=$1", id).Scan(&o.ID, &o.CustomerID, &o.Amount, &o.Recent, &o.Status)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		server.JSON(w, 404, map[string]string{"error": "order not found"})
	} else if err != nil {
		server.JSON(w, 503, map[string]string{"error": "database unavailable"})
	} else {
		server.JSON(w, 200, o)
	}
}
