package telemetry

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const Schema = "1"

type Window struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

func (w Window) Validate() error {
	if w.Start.IsZero() || !w.End.After(w.Start) || w.End.Sub(w.Start) < 30*time.Second || w.End.Sub(w.Start) > 24*time.Hour {
		return fmt.Errorf("window must be between 30 seconds and 24 hours")
	}
	return nil
}

type Target struct {
	Service   string `json:"service"`
	Route     string `json:"route"`
	Namespace string `json:"namespace"`
}

type Measurement struct {
	Value   *float64 `json:"value"`
	Unit    string   `json:"unit"`
	Query   string   `json:"query"`
	Error   string   `json:"error,omitempty"`
	Samples int      `json:"samples"`
}

type TraceSummary struct {
	ID             string             `json:"id"`
	RootDuration   float64            `json:"root_duration_seconds"`
	QuerySeconds   float64            `json:"query_seconds"`
	AcquireSeconds float64            `json:"acquire_seconds"`
	LocalSeconds   float64            `json:"local_seconds"`
	Operations     map[string]float64 `json:"operation_seconds"`
	Incomplete     bool               `json:"incomplete"`
}

type Snapshot struct {
	Schema          string                     `json:"schema_version"`
	CapturedAt      time.Time                  `json:"captured_at"`
	Target          Target                     `json:"target"`
	Window          Window                     `json:"window"`
	Metrics         map[string]Measurement     `json:"metrics"`
	Traces          []TraceSummary             `json:"traces"`
	TraceQuery      string                     `json:"trace_query"`
	TraceCandidates int                        `json:"trace_candidates"`
	TraceIncomplete int                        `json:"trace_incomplete"`
	TraceLimit      int                        `json:"trace_limit"`
	Sampling        string                     `json:"sampling"`
	Missing         []string                   `json:"missing"`
	Workload        string                     `json:"workload,omitempty"`
	OfferedRate     *float64                   `json:"offered_rate,omitempty"`
	Healthy         bool                       `json:"healthy"`
	SLOSeconds      float64                    `json:"slo_seconds"`
	Source          map[string]string          `json:"source"`
	RawTraces       map[string]json.RawMessage `json:"raw_traces,omitempty"`
}

func (s Snapshot) Value(key string) (float64, bool) {
	m, ok := s.Metrics[key]
	if !ok || m.Value == nil {
		return 0, false
	}
	return *m.Value, true
}

func Load(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

func Save(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}
