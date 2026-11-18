package diagnosis

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/panjitamzil/scalens/internal/telemetry"
)

type Rules struct {
	MinRequests     float64 `json:"min_requests"`
	LatencyIncrease float64 `json:"latency_increase"`
	ErrorRate       float64 `json:"error_rate"`
	SLOBadRate      float64 `json:"slo_bad_rate"`
	CPUFraction     float64 `json:"cpu_fraction"`
	QuerySeconds    float64 `json:"query_seconds"`
	ServerSeconds   float64 `json:"server_query_seconds"`
}

func DefaultRules() Rules { return Rules{30, 1.5, .01, .05, .6, .05, .01} }

type Hypothesis struct {
	Category    string   `json:"category"`
	Service     string   `json:"service"`
	Operations  []string `json:"operations,omitempty"`
	Evidence    []string `json:"evidence"`
	Limitations []string `json:"limitations"`
	Experiments []string `json:"experiments"`
}

type Report struct {
	Schema       string              `json:"schema_version"`
	CreatedAt    time.Time           `json:"created_at"`
	Status       string              `json:"status"`
	Observations []string            `json:"observations"`
	Suspects     []Hypothesis        `json:"suspects"`
	Missing      []string            `json:"missing_evidence"`
	Warnings     []string            `json:"warnings"`
	Rules        Rules               `json:"rules"`
	Snapshot     telemetry.Snapshot  `json:"snapshot"`
	Baseline     *telemetry.Snapshot `json:"baseline,omitempty"`
}

func CheckBaseline(s, b telemetry.Snapshot) error {
	if b.Schema != telemetry.Schema || s.Target != b.Target {
		return fmt.Errorf("baseline schema or target does not match")
	}
	if !b.Healthy {
		return fmt.Errorf("baseline is not marked healthy")
	}
	if b.Window.Start.Before(s.Window.End) && s.Window.Start.Before(b.Window.End) {
		return fmt.Errorf("baseline overlaps observation window")
	}
	if s.Workload != "" && b.Workload != "" && s.Workload != b.Workload {
		return fmt.Errorf("baseline workload identity differs")
	}
	return nil
}

func Evaluate(s telemetry.Snapshot, b *telemetry.Snapshot, rules Rules) (Report, error) {
	report := Report{Schema: telemetry.Schema, CreatedAt: time.Now().UTC(), Status: "insufficient_evidence", Rules: rules, Snapshot: s, Baseline: b, Observations: []string{}, Suspects: []Hypothesis{}, Warnings: []string{}, Missing: append([]string{}, s.Missing...)}
	if s.Schema != telemetry.Schema {
		return report, fmt.Errorf("unsupported snapshot schema %q", s.Schema)
	}
	for _, v := range []float64{rules.MinRequests, rules.LatencyIncrease, rules.ErrorRate, rules.SLOBadRate, rules.CPUFraction, rules.QuerySeconds, rules.ServerSeconds} {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
			return report, fmt.Errorf("rule thresholds must be finite and nonnegative")
		}
	}
	if rules.MinRequests < 1 || rules.LatencyIncrease <= 1 || rules.CPUFraction > 1 || rules.ErrorRate > 1 || rules.SLOBadRate > 1 {
		return report, fmt.Errorf("invalid rule thresholds")
	}
	if b != nil {
		if err := CheckBaseline(s, *b); err != nil {
			return report, err
		}
	}
	n, ok := s.Value("requests")
	if !ok || n < rules.MinRequests {
		report.Missing = append(report.Missing, "insufficient observed requests")
		return report, nil
	}
	lat, hasLat := s.Value("success_p95")
	errRate, hasError := s.Value("error_rate")
	bad, hasSLO := s.Value("slo_bad_rate")
	successes, hasSuccesses := s.Value("successes")
	if (hasSuccesses && successes > n*1.02) || (hasSLO && hasError && bad+.02 < errRate) {
		report.Status = "conflicting_evidence"
		report.Observations = append(report.Observations, "Request, success or SLO counters disagree; check scope, reset and scrape boundaries")
		return report, nil
	}
	symptom := false
	comparison := false
	if hasError && errRate > rules.ErrorRate {
		symptom = true
		report.Observations = append(report.Observations, fmt.Sprintf("Error rate %.3f exceeds configured %.3f", errRate, rules.ErrorRate))
	}
	if s.SLOSeconds > 0 && hasSLO {
		comparison = true
		if bad > rules.SLOBadRate {
			symptom = true
			report.Observations = append(report.Observations, fmt.Sprintf("SLO bad-event rate %.3f exceeds %.3f", bad, rules.SLOBadRate))
		}
	}
	if b != nil {
		if base, ok := b.Value("success_p95"); ok && hasLat && base > 0 {
			comparison = true
			if lat > base*rules.LatencyIncrease {
				symptom = true
				report.Observations = append(report.Observations, fmt.Sprintf("Successful p95 increased from %.4fs to %.4fs", base, lat))
			}
		}
		if s.Workload != "" && s.Workload == b.Workload && s.OfferedRate != nil && b.OfferedRate != nil && *s.OfferedRate == *b.OfferedRate {
			before, bok := b.Value("success_rps")
			after, aok := s.Value("success_rps")
			if bok && aok && before > 0 && after < before*.8 {
				symptom = true
				report.Observations = append(report.Observations, "Successful throughput fell by more than 20% at the same offered rate")
			}
		} else {
			report.Warnings = append(report.Warnings, "Throughput regression not inferred without matched offered-load metadata")
		}
	}
	if !comparison {
		report.Missing = append(report.Missing, "no usable baseline comparison or SLO measurement")
		return report, nil
	}
	health, hasHealth := s.Value("scrape_health")
	covered := hasHealth && health >= .95
	if !covered {
		report.Missing = append(report.Missing, "application scrape coverage below 95% or unknown")
	}
	if !symptom {
		if !covered {
			return report, nil
		}
		if hasLat && hasError {
			report.Status = "no_anomaly"
			report.Observations = append(report.Observations, "No degradation detected in the available latency, error and SLO signals")
		} else {
			report.Missing = append(report.Missing, "latency or error evidence missing")
		}
		return report, nil
	}
	if frac, ok := s.Value("cpu_near_limit_fraction"); ok && frac >= rules.CPUFraction {
		h := Hypothesis{
			Category: "cpu_saturation", Service: s.Target.Service,
			Evidence:    []string{"cpu_ratio", "cpu_near_limit_fraction"},
			Limitations: []string{"Container CPU evidence does not identify a CPU-expensive function", "Host contention must be checked independently"},
			Experiments: []string{"Compare one and two replicas at the same offered load with host headroom", "Inspect CPU profiles if local computation remains unclear"},
		}
		if _, ok := s.Value("cpu_throttled_ratio"); ok {
			h.Evidence = append(h.Evidence, "cpu_throttled_ratio")
		} else {
			h.Limitations = append(h.Limitations, "CFS throttling counters unavailable")
		}
		report.Suspects = append(report.Suspects, h)
	}
	cpuHot := len(report.Suspects) > 0
	q, qok := s.Value("query_p95")
	queryThreshold := rules.QuerySeconds
	if b != nil {
		if base, ok := b.Value("query_p95"); ok && base*rules.LatencyIncrease > queryThreshold {
			queryThreshold = base * rules.LatencyIncrease
		}
	}
	server, hasServer := s.Value("db_server_mean")
	switch {
	case !qok || q <= queryThreshold:
	case hasServer && server < rules.ServerSeconds:
		report.Observations = append(report.Observations, fmt.Sprintf("Client-side database p95 is %.3fs but the costliest statement averages %.4fs inside PostgreSQL; the time is spent outside query execution", q, server))
	case !hasServer && cpuHot:
		report.Missing = append(report.Missing, "database timing was measured inside a CPU-saturated process and server-side statement statistics are unavailable")
	default:
		h := Hypothesis{Category: "slow_database_operation", Service: s.Target.Service, Evidence: []string{"query_p95"}, Limitations: []string{"Statement time includes execution, lock and I/O waiting", "Neither a missing index nor PostgreSQL saturation is proven by slow statements"}, Experiments: []string{"Inspect EXPLAIN (ANALYZE, BUFFERS) and pg_stat_statements for the operation", "Compare the relevant index change with application scale-out"}}
		if hasServer {
			h.Evidence = append(h.Evidence, "db_server_mean")
		} else {
			h.Limitations = append(h.Limitations, "No server-side statement statistics; client timing includes network and client scheduling")
		}
		ops := map[string]float64{}
		for _, tr := range s.Traces {
			slow := false
			for op, seconds := range tr.Operations {
				if seconds >= rules.QuerySeconds {
					ops[op] += seconds
					slow = true
				}
			}
			if slow {
				h.Evidence = append(h.Evidence, "trace:"+tr.ID)
			}
		}
		for op := range ops {
			h.Operations = append(h.Operations, op)
		}
		sort.Slice(h.Operations, func(i, j int) bool { return ops[h.Operations[i]] > ops[h.Operations[j]] })
		if len(h.Operations) > 1 {
			h.Limitations = append(h.Limitations, "Operations are ordered by sampled time; a saturated database slows unrelated statements too")
		}
		if len(h.Operations) == 0 {
			h.Limitations = append(h.Limitations, "No sampled trace corroborates the metric")
		}
		if acq, ok := s.Value("acquire_p95"); ok && acq > q {
			h.Limitations = append(h.Limitations, "Connection acquisition is slower than query execution; pool waiting may also contribute")
		}
		report.Suspects = append(report.Suspects, h)
	}
	if len(report.Suspects) > 0 {
		report.Status = "suspected"
	} else {
		report.Missing = append(report.Missing, "symptoms observed, but neither CPU nor database rule has enough supporting evidence")
	}
	if !covered {
		report.Warnings = append(report.Warnings, "Scrapes were missed during the window; counters are under-sampled and an overloaded target can cause this")
	}
	if len(report.Suspects) > 1 {
		report.Warnings = append(report.Warnings, "Multiple supported suspects; evaluate interventions separately")
	}
	report.Warnings = append(report.Warnings, "Sampled traces are selected evidence, not a population latency estimate")
	return report, nil
}
