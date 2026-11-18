package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMissingAndNonFiniteAreNotZero(t *testing.T) {
	for _, body := range []string{
		`{"status":"success","data":{"result":[]}}`,
		`{"status":"success","data":{"result":[{"value":[1,"NaN"]}]}}`,
		`{"status":"error","error":"query failed"}`,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) }))
		c := New(Config{Prometheus: server.URL})
		m := c.metric(context.Background(), "query", "seconds", Window{End: time.Now()})
		server.Close()
		if m.Value != nil || m.Error == "" {
			t.Fatalf("missing data became a value: %+v", m)
		}
	}
}
