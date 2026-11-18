package diagnosis

import (
	"testing"
	"time"

	"github.com/panjitamzil/scalens/internal/telemetry"
)

func snapshot(values map[string]float64) telemetry.Snapshot {
	s := telemetry.Snapshot{Schema: "1", SLOSeconds: .5, Target: telemetry.Target{Service: "order-service", Route: "/orders", Namespace: "lab"}, Window: telemetry.Window{Start: time.Unix(1000, 0), End: time.Unix(1060, 0)}, Metrics: map[string]telemetry.Measurement{}}
	for key, v := range values {
		s.Metrics[key] = telemetry.Measurement{Value: &v}
	}
	return s
}

func TestDiagnosisOutcomes(t *testing.T) {
	const cpu, db = "cpu_saturation", "slow_database_operation"
	tests := []struct {
		name    string
		set     map[string]float64
		drop    string
		noSLO   bool
		status  string
		suspect []string
	}{
		{name: "healthy", status: "no_anomaly"},
		{name: "CPU", set: map[string]float64{"slo_bad_rate": .8, "cpu_near_limit_fraction": .8}, status: "suspected", suspect: []string{cpu}},
		{name: "database", set: map[string]float64{"slo_bad_rate": .8, "query_p95": .8, "db_server_mean": .2}, status: "suspected", suspect: []string{db}},
		{name: "database without server statistics", set: map[string]float64{"slo_bad_rate": .8, "query_p95": .8}, status: "suspected", suspect: []string{db}},
		{name: "two supported suspects", set: map[string]float64{"slo_bad_rate": .8, "query_p95": .8, "db_server_mean": .2, "cpu_near_limit_fraction": .8}, status: "suspected", suspect: []string{cpu, db}},
		{name: "CPU starvation inflates client query time", set: map[string]float64{"slo_bad_rate": .8, "query_p95": .8, "db_server_mean": .0002, "cpu_near_limit_fraction": .8}, status: "suspected", suspect: []string{cpu}},
		{name: "CPU hot and no server statistics", set: map[string]float64{"slo_bad_rate": .8, "query_p95": .8, "cpu_near_limit_fraction": .8}, status: "suspected", suspect: []string{cpu}},
		{name: "slow client query with fast server", set: map[string]float64{"slo_bad_rate": .8, "query_p95": .8, "db_server_mean": .0002}, status: "insufficient_evidence"},
		{name: "missing request metrics", drop: "requests", status: "insufficient_evidence"},
		{name: "no baseline or SLO", noSLO: true, status: "insufficient_evidence"},
		{name: "missing scrape coverage", drop: "scrape_health", status: "insufficient_evidence"},
		{name: "missed scrapes do not hide a supported suspect", set: map[string]float64{"slo_bad_rate": .8, "cpu_near_limit_fraction": 1, "scrape_health": .6}, status: "suspected", suspect: []string{cpu}},
		{name: "conflicting counters", set: map[string]float64{"error_rate": .8}, status: "conflicting_evidence"},
		{name: "symptom without cause evidence", set: map[string]float64{"slo_bad_rate": .8}, status: "insufficient_evidence"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := snapshot(map[string]float64{"requests": 100, "successes": 100, "success_p95": .02, "error_rate": 0, "slo_bad_rate": 0, "scrape_health": 1})
			for key, v := range tc.set {
				s.Metrics[key] = telemetry.Measurement{Value: &v}
			}
			delete(s.Metrics, tc.drop)
			if tc.noSLO {
				s.SLOSeconds = 0
			}
			r, err := Evaluate(s, nil, DefaultRules())
			if err != nil || r.Status != tc.status || len(r.Suspects) != len(tc.suspect) {
				t.Fatalf("status=%s suspects=%+v missing=%v err=%v", r.Status, r.Suspects, r.Missing, err)
			}
			for i, want := range tc.suspect {
				if r.Suspects[i].Category != want {
					t.Fatalf("suspect %d is %s, want %s", i, r.Suspects[i].Category, want)
				}
			}
		})
	}
}

func TestBaselineComparison(t *testing.T) {
	base := snapshot(map[string]float64{"requests": 100, "successes": 100, "success_p95": .02, "error_rate": 0, "scrape_health": 1, "query_p95": .004})
	base.SLOSeconds, base.Healthy = 0, true
	base.Window = telemetry.Window{Start: time.Unix(500, 0), End: time.Unix(560, 0)}
	s := snapshot(map[string]float64{"requests": 100, "successes": 100, "success_p95": .2, "error_rate": 0, "scrape_health": 1, "query_p95": .15, "db_server_mean": .12})
	s.SLOSeconds = 0
	r, err := Evaluate(s, &base, DefaultRules())
	if err != nil || r.Status != "suspected" || len(r.Suspects) != 1 || r.Suspects[0].Category != "slow_database_operation" {
		t.Fatalf("baseline regression not diagnosed: %+v %v", r, err)
	}
	s.Metrics["success_p95"] = base.Metrics["success_p95"]
	if r, _ = Evaluate(s, &base, DefaultRules()); r.Status != "no_anomaly" {
		t.Fatalf("unchanged latency reported as %s", r.Status)
	}
}

func TestBaselineGuards(t *testing.T) {
	s := snapshot(nil)
	b := s
	b.Healthy = true
	if CheckBaseline(s, b) == nil {
		t.Fatal("overlapping baseline accepted")
	}
	b.Window = telemetry.Window{Start: time.Unix(500, 0), End: time.Unix(560, 0)}
	if err := CheckBaseline(s, b); err != nil {
		t.Fatal(err)
	}
	b.Healthy = false
	if CheckBaseline(s, b) == nil {
		t.Fatal("unhealthy baseline accepted")
	}
	b.Healthy = true
	b.Target.Route = "/payments"
	if CheckBaseline(s, b) == nil {
		t.Fatal("different endpoint accepted")
	}
}
