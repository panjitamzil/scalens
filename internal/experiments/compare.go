package experiments

import (
	"fmt"
	"math"
	"path/filepath"
	"sort"

	"github.com/panjitamzil/scalens/internal/telemetry"
)

type Stats struct {
	Requests   float64  `json:"requests"`
	Successes  float64  `json:"successes"`
	Duration   float64  `json:"duration_seconds"`
	ErrorRate  float64  `json:"error_rate"`
	SLOBadRate float64  `json:"slo_bad_rate"`
	P95        *float64 `json:"success_p95_ms"`
	P99        *float64 `json:"success_p99_ms"`
	Dropped    float64  `json:"dropped_iterations"`
}

type Run struct {
	Schema          string   `json:"schema_version"`
	ID              string   `json:"id"`
	Scenario        string   `json:"scenario"`
	Arm             string   `json:"arm"`
	Block           int      `json:"block"`
	Rate            float64  `json:"rate"`
	Seconds         int      `json:"seconds"`
	Rows            int      `json:"rows"`
	Profile         string   `json:"profile"`
	ResourceProfile string   `json:"resource_profile"`
	Validity        string   `json:"validity"`
	Reasons         []string `json:"reasons"`
	Start           string   `json:"start"`
	End             string   `json:"end"`
	Stats           Stats    `json:"stats"`
}

type Aggregate struct {
	Runs       int        `json:"runs"`
	P95Median  float64    `json:"success_p95_median_ms"`
	P95Range   [2]float64 `json:"success_p95_range_ms"`
	SuccessRPS float64    `json:"successful_rps_median"`
	ErrorRate  float64    `json:"error_rate_median"`
	SLOBadRate float64    `json:"slo_bad_rate_median"`
}

type Comparison struct {
	Schema           string    `json:"schema_version"`
	Status           string    `json:"status"`
	Quality          string    `json:"evidence_quality"`
	Control          Aggregate `json:"control"`
	Candidate        Aggregate `json:"candidate"`
	P95Change        float64   `json:"p95_relative_change"`
	ThroughputChange float64   `json:"throughput_relative_change"`
	Excluded         []string  `json:"excluded"`
	Notes            []string  `json:"notes"`
	ControlRuns      []Run     `json:"control_runs"`
	CandidateRuns    []Run     `json:"candidate_runs"`
}

func ReadRuns(dir string) ([]Run, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*", "run.json"))
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		paths, err = filepath.Glob(filepath.Join(dir, "run.json"))
		if err != nil {
			return nil, err
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no run.json files in %s", dir)
	}
	var runs []Run
	for _, p := range paths {
		var r Run
		if err := telemetry.Load(p, &r); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, nil
}

func median(v []float64) float64 {
	sort.Float64s(v)
	n := len(v)
	if n%2 == 1 {
		return v[n/2]
	}
	return (v[n/2-1] + v[n/2]) / 2
}

func aggregate(runs []Run) Aggregate {
	var p95, rps, errors, bad []float64
	for _, r := range runs {
		p95 = append(p95, *r.Stats.P95)
		rps = append(rps, r.Stats.Successes/r.Stats.Duration)
		errors = append(errors, r.Stats.ErrorRate)
		bad = append(bad, r.Stats.SLOBadRate)
	}
	m := median(p95)
	return Aggregate{len(runs), m, [2]float64{p95[0], p95[len(p95)-1]}, median(rps), median(errors), median(bad)}
}

func Compare(control, candidate []Run, includeFlagged bool) (Comparison, error) {
	out := Comparison{Schema: telemetry.Schema, Status: "inconclusive", Quality: "valid_runs_only", Notes: []string{"Exploratory comparison; no statistical significance is claimed", "A diagnosis can be correct even when the intervention is ineffective", "Practical thresholds: 10% latency, 5% throughput, 1 percentage point error/SLO guardrail; 5 points when both arms fail most requests"}}
	all := append(append([]Run{}, control...), candidate...)
	if len(control) == 0 || len(candidate) == 0 {
		return out, fmt.Errorf("both arms need runs")
	}
	reference := control[0]
	for _, r := range all {
		if r.Schema != telemetry.Schema || r.Scenario != reference.Scenario || r.Rate != reference.Rate || r.Rows != reference.Rows || r.Seconds != reference.Seconds || r.Profile != reference.Profile || r.ResourceProfile != reference.ResourceProfile {
			return out, fmt.Errorf("incompatible run %s; workload, resource profile and duration must match", r.ID)
		}
	}
	filter := func(runs []Run) []Run {
		var valid []Run
		for _, r := range runs {
			flagged := r.Validity == "contended" || r.Validity == "load_not_achieved" || r.Stats.Dropped > 0
			if r.Stats.P95 == nil || r.Stats.Duration <= 0 || r.Stats.Successes <= 0 || (flagged && !includeFlagged) || (!flagged && r.Validity != "valid") {
				out.Excluded = append(out.Excluded, r.ID+": "+r.Validity)
				continue
			}
			if flagged && out.Quality != "includes_flagged_runs" {
				out.Quality = "includes_flagged_runs"
				out.Notes = append(out.Notes, "Runs flagged for host contention or dropped iterations are included; treat the outcome as indicative and repeat on a machine with headroom", "Where k6 dropped iterations the offered load was not reached and latency is understated")
			}
			valid = append(valid, r)
		}
		return valid
	}
	control = filter(control)
	candidate = filter(candidate)
	out.ControlRuns = control
	out.CandidateRuns = candidate
	if len(control) == 0 || len(candidate) == 0 {
		out.Notes = append(out.Notes, "No usable data in at least one arm")
		return out, nil
	}
	out.Control = aggregate(control)
	out.Candidate = aggregate(candidate)
	if out.Control.P95Median <= 0 || out.Control.SuccessRPS <= 0 {
		return out, nil
	}
	out.P95Change = (out.Candidate.P95Median - out.Control.P95Median) / out.Control.P95Median
	out.ThroughputChange = (out.Candidate.SuccessRPS - out.Control.SuccessRPS) / out.Control.SuccessRPS
	if len(control) < 3 || len(candidate) < 3 {
		out.Notes = append(out.Notes, "Fewer than three valid runs per arm")
		return out, nil
	}
	// Latency of a few survivors says little when both arms fail most requests.
	if out.Control.ErrorRate > .5 && out.Candidate.ErrorRate > .5 {
		out.Notes = append(out.Notes, "Both arms failed most requests; only error and SLO rates are compared")
		if math.Abs(out.Candidate.ErrorRate-out.Control.ErrorRate) < .05 && math.Abs(out.Candidate.SLOBadRate-out.Control.SLOBadRate) < .05 {
			out.Status = "no_material_change_observed"
		} else if out.Candidate.ErrorRate < out.Control.ErrorRate {
			out.Status = "improved"
		} else {
			out.Status = "regressed"
		}
		return out, nil
	}
	good := out.P95Change <= -.1 || out.ThroughputChange >= .05
	bad := out.P95Change >= .1 || out.ThroughputChange <= -.05 || out.Candidate.ErrorRate > out.Control.ErrorRate+.01 || out.Candidate.SLOBadRate > out.Control.SLOBadRate+.01
	switch {
	case good && bad:
		out.Status = "mixed"
	case bad:
		out.Status = "regressed"
	case good:
		out.Status = "improved"
	case math.Abs(out.P95Change) < .1 && math.Abs(out.ThroughputChange) < .05:
		out.Status = "no_material_change_observed"
	}
	return out, nil
}
