package experiments

import "testing"

func runs(p95, successes, errors float64) []Run {
	var result []Run
	for i := 0; i < 3; i++ {
		result = append(result, Run{Schema: "1", ID: "test", Scenario: "cpu", Rate: 10, Seconds: 60, Rows: 1000, Profile: "constant", ResourceProfile: "test", Validity: "valid", Stats: Stats{Duration: 60, Requests: 600, Successes: successes, P95: &p95, ErrorRate: errors}})
	}
	return result
}

func TestFastFailuresAreNotImprovement(t *testing.T) {
	r, err := Compare(runs(500, 600, 0), runs(100, 200, .66), false)
	if err != nil || r.Status != "mixed" {
		t.Fatalf("fast failures presented as improvement: %+v, %v", r, err)
	}
}

func TestInvalidAndMismatchedRuns(t *testing.T) {
	a, b := runs(500, 600, 0), runs(100, 600, 0)
	b[0].Validity = "contended"
	r, err := Compare(a, b, false)
	if err != nil || r.Status != "inconclusive" || len(r.Excluded) != 1 {
		t.Fatalf("invalid run included: %+v %v", r, err)
	}
	r, err = Compare(a, b, true)
	if err != nil || r.Status != "improved" || r.Quality != "includes_flagged_runs" {
		t.Fatalf("contended runs not labelled: %+v %v", r, err)
	}
	b[0].Rate = 20
	if _, err := Compare(a, b, false); err == nil {
		t.Fatal("mismatched workload accepted")
	}
}

func TestBothArmsFailing(t *testing.T) {
	r, err := Compare(runs(5000, 6, .99), runs(7000, 2, .995), false)
	if err != nil || r.Status != "no_material_change_observed" {
		t.Fatalf("noise between failing arms reported as %s (%v)", r.Status, err)
	}
}
