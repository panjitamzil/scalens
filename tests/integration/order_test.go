package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/panjitamzil/scalens/internal/observability"
	"github.com/panjitamzil/scalens/services/order"
	"github.com/panjitamzil/scalens/services/payment"
)

func TestOrderPersistenceAndPayment(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is required")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	for _, tc := range []struct {
		name    string
		fail    bool
		delay   time.Duration
		timeout time.Duration
		code    int
		status  string
	}{
		{"paid", false, 0, time.Second, 201, "paid"},
		{"failed", true, 0, time.Second, 502, "failed"},
		{"unknown", false, 200 * time.Millisecond, 20 * time.Millisecond, 504, "payment_unknown"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := httptest.NewServer(payment.Handler(observability.NewMetrics("payment-service"), tc.delay, tc.fail))
			defer p.Close()
			svc := order.New(pool, observability.NewMetrics("order-service"), p.URL, 0)
			svc.Client.Timeout = tc.timeout
			s := httptest.NewServer(svc.Handler())
			defer s.Close()
			resp, err := http.Post(s.URL+"/orders", "application/json", bytes.NewBufferString(`{"customer_id":1,"items":[{"price":1000,"quantity":2}]}`))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			var o order.Order
			if err := json.NewDecoder(resp.Body).Decode(&o); err != nil {
				t.Fatal(err)
			}
			var status string
			if err := pool.QueryRow(context.Background(), "SELECT status FROM orders WHERE id=$1", o.ID).Scan(&status); err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.code || status != tc.status || o.Amount != 2200 {
				t.Fatalf("response=%d order=%+v persisted=%s", resp.StatusCode, o, status)
			}
		})
	}
}
