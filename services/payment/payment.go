package payment

import (
	"context"
	"net/http"
	"time"

	"github.com/panjitamzil/scalens/internal/observability"
	"github.com/panjitamzil/scalens/internal/server"
	"go.opentelemetry.io/otel"
)

type Request struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

func Handler(m *observability.Metrics, delay time.Duration, fail bool) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /payments", m.Wrap("/payments", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer("scalens/payment").Start(r.Context(), "PaymentService.ProcessPayment")
		defer span.End()
		var req Request
		if err := server.Decode(w, r, &req); err != nil || req.OrderID == "" || req.Amount <= 0 {
			server.JSON(w, 400, map[string]string{"error": "invalid payment"})
			return
		}
		if err := wait(ctx, delay); err != nil {
			span.RecordError(err)
			server.JSON(w, 504, map[string]string{"error": "payment canceled"})
			return
		}
		if fail {
			server.JSON(w, 503, map[string]string{"error": "payment unavailable"})
			return
		}
		server.JSON(w, 200, map[string]string{"status": "paid", "order_id": req.OrderID})
	}))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	mux.Handle("GET /metrics", m.Handler())
	return mux
}

func wait(ctx context.Context, delay time.Duration) error {
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
