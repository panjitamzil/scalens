package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/panjitamzil/scalens/internal/telemetry"
)

func TestWindowFlags(t *testing.T) {
	for _, args := range [][]string{
		{"diagnose", "--last", "5m", "--start", "2026-11-20T10:00:00Z"},
		{"diagnose", "--start", "2026-11-20T10:00:00Z"},
		{"diagnose", "--start", "2026-11-20T10:00:00Z", "--end", "2026-11-20T10:00:05Z"},
		{"diagnose", "--start", "yesterday", "--end", "2026-11-20T10:00:05Z"},
		{"diagnose", "--input", "x.json", "--last", "5m"},
		{"baseline", "capture", "--last", "5m", "--baseline-file", "x.json"},
		{"scale"},
	} {
		if err := run(args, &bytes.Buffer{}); err == nil {
			t.Errorf("%v accepted", args)
		}
	}
}

// The fake backends answer every query with one healthy sample and no traces.
func TestLiveCaptureThenReplay(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/query_range"):
			w.Write([]byte(`{"status":"success","data":{"result":[{"values":[[1,"0.1"],[2,"0.1"],[3,"0.1"]]}]}}`))
		case strings.HasSuffix(r.URL.Path, "/query"):
			value := "0.02"
			q := r.URL.Query().Get("query")
			if strings.HasPrefix(q, "sum(increase(http_server_requests_total") || strings.HasPrefix(q, "count(") || strings.HasPrefix(q, "min(avg_over_time(up") {
				value = "1"
			}
			if strings.HasPrefix(q, "sum(increase(http_server_requests_total") {
				value = "200"
			}
			if strings.HasPrefix(q, "(sum(increase(http_server_requests_total") && strings.Contains(q, " - ") || strings.HasPrefix(q, "1 - ") {
				value = "0"
			}
			w.Write([]byte(`{"status":"success","data":{"result":[{"value":[1,"` + value + `"]}]}}`))
		default:
			w.Write([]byte(`{"traces":[]}`))
		}
	}))
	defer backend.Close()
	dir := t.TempDir()
	config := filepath.Join(dir, "config.json")
	raw, _ := json.Marshal(map[string]any{"prometheus": backend.URL, "tempo": backend.URL, "namespace": "lab", "trace_limit": 3, "cpu_near_limit": .8})
	if err := os.WriteFile(config, raw, 0600); err != nil {
		t.Fatal(err)
	}
	baseline := filepath.Join(dir, "healthy.json")
	args := []string{"baseline", "capture", "--config", config, "--start", "2026-11-20T10:00:00Z", "--end", "2026-11-20T10:05:00Z", "--out", baseline}
	if err := run(args, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var saved telemetry.Snapshot
	if err := telemetry.Load(baseline, &saved); err != nil || !saved.Healthy || saved.Target.Route != "/orders" || saved.Target.Namespace != "lab" {
		t.Fatalf("baseline metadata: %+v %v", saved, err)
	}
	var out bytes.Buffer
	args = []string{"diagnose", "--config", config, "--start", "2026-11-20T10:10:00Z", "--end", "2026-11-20T10:15:00Z", "--baseline-file", baseline, "--format", "json"}
	if err := run(args, &out); err != nil {
		t.Fatal(err)
	}
	var report struct {
		Status  string   `json:"status"`
		Missing []string `json:"missing_evidence"`
	}
	if err := json.Unmarshal(out.Bytes(), &report); err != nil || report.Status != "no_anomaly" {
		t.Fatalf("status=%q err=%v", report.Status, err)
	}
	if !strings.Contains(strings.Join(report.Missing, "\n"), "no traces found") {
		t.Fatalf("missing traces not reported: %v", report.Missing)
	}
	args[5], args[7] = "2026-11-20T10:04:00Z", "2026-11-20T10:09:00Z"
	if err := run(args, &bytes.Buffer{}); err == nil {
		t.Fatal("overlapping baseline accepted")
	}
	out.Reset()
	if err := run([]string{"diagnose", "--input", baseline}, &out); err != nil || !strings.Contains(out.String(), "**no_anomaly**") {
		t.Fatalf("replay failed: %v\n%s", err, out.String())
	}
}

func TestBackendDownIsNotHealthy(t *testing.T) {
	backend := httptest.NewServer(http.NotFoundHandler())
	backend.Close()
	dir := t.TempDir()
	config := filepath.Join(dir, "config.json")
	raw, _ := json.Marshal(map[string]any{"prometheus": backend.URL, "tempo": backend.URL, "namespace": "lab", "trace_limit": 3, "cpu_near_limit": .8})
	if err := os.WriteFile(config, raw, 0600); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := run([]string{"diagnose", "--config", config, "--last", "5m", "--format", "json"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"status": "insufficient_evidence"`) {
		t.Fatalf("unreachable backends produced: %s", out.String())
	}
}
