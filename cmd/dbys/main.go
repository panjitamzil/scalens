package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/panjitamzil/scalens/internal/diagnosis"
	"github.com/panjitamzil/scalens/internal/reporting"
	"github.com/panjitamzil/scalens/internal/telemetry"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "dbys:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: dbys baseline capture | diagnose | compare")
	}
	if args[0] == "compare" {
		return compareCommand(args[1:], out)
	}
	capture := args[0] == "baseline" && len(args) > 1 && args[1] == "capture"
	if args[0] != "diagnose" && !capture {
		return errors.New("unknown command")
	}
	if capture {
		args = args[2:]
	} else {
		args = args[1:]
	}
	fs := flag.NewFlagSet("dbys", flag.ContinueOnError)
	service := fs.String("service", "order-service", "service name")
	route := fs.String("endpoint", "POST /orders", "method and route")
	start := fs.String("start", "", "RFC3339 start")
	end := fs.String("end", "", "RFC3339 end")
	last := fs.Duration("last", 0, "observation duration ending now")
	config := fs.String("config", "", "JSON backend and rule configuration")
	baseFile := fs.String("baseline-file", "", "captured healthy baseline")
	baseStart := fs.String("baseline-start", "", "RFC3339 baseline start")
	baseEnd := fs.String("baseline-end", "", "RFC3339 baseline end")
	input := fs.String("input", "", "replay a telemetry snapshot")
	output := fs.String("out", "", "report/baseline file")
	snapshotOut := fs.String("snapshot-out", "", "save live telemetry snapshot")
	format := fs.String("format", "markdown", "markdown or json")
	workload := fs.String("workload", "", "workload identity")
	rate := fs.Float64("offered-rate", 0, "known offered requests per second")
	slo := fs.Float64("slo", .5, "latency SLO in seconds; 0 disables")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	if *slo < 0 || *rate < 0 {
		return errors.New("slo and offered-rate cannot be negative")
	}
	var cfg struct {
		telemetry.Config
		Rules diagnosis.Rules `json:"rules"`
	}
	cfg.Config = telemetry.DefaultConfig()
	cfg.Rules = diagnosis.DefaultRules()
	if *config != "" {
		if err := telemetry.Load(*config, &cfg); err != nil {
			return err
		}
	}
	var s telemetry.Snapshot
	if *input != "" {
		if capture || *start != "" || *end != "" || *last != 0 {
			return errors.New("input cannot be combined with capture or time flags")
		}
		if err := telemetry.Load(*input, &s); err != nil {
			return err
		}
	} else {
		w, err := window(*start, *end, *last)
		if err != nil {
			return err
		}
		var method, path string
		if _, err := fmt.Sscanf(*route, "%s %s", &method, &path); err != nil || method != "POST" {
			return errors.New("MVP endpoint must be POST and a route template")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		s, err = telemetry.New(cfg.Config).Collect(ctx, telemetry.Target{Service: *service, Route: path, Namespace: cfg.Namespace}, w, *slo)
		if err != nil {
			return err
		}
		s.Workload = *workload
		if *rate > 0 {
			s.OfferedRate = rate
		}
	}
	if *snapshotOut != "" {
		if err := telemetry.Save(*snapshotOut, s); err != nil {
			return err
		}
	}
	if capture {
		if *baseFile != "" || *baseStart != "" || *baseEnd != "" {
			return errors.New("capture does not accept a baseline")
		}
		r, err := diagnosis.Evaluate(s, nil, cfg.Rules)
		if err != nil {
			return err
		}
		if r.Status != "no_anomaly" {
			return fmt.Errorf("baseline is not healthy: %s; inspect a diagnosis report", r.Status)
		}
		s.Healthy = true
		if *output == "" {
			return errors.New("baseline capture requires --out")
		}
		return telemetry.Save(*output, s)
	}
	var base *telemetry.Snapshot
	if *baseFile != "" {
		if *baseStart != "" || *baseEnd != "" {
			return errors.New("choose a baseline file or window")
		}
		base = &telemetry.Snapshot{}
		if err := telemetry.Load(*baseFile, base); err != nil {
			return err
		}
	} else if *baseStart != "" || *baseEnd != "" {
		bw, err := window(*baseStart, *baseEnd, 0)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		b, err := telemetry.New(cfg.Config).Collect(ctx, s.Target, bw, *slo)
		if err != nil {
			return err
		}
		health, err := diagnosis.Evaluate(b, nil, cfg.Rules)
		if err != nil {
			return err
		}
		b.Healthy = health.Status == "no_anomaly"
		base = &b
	}
	r, err := diagnosis.Evaluate(s, base, cfg.Rules)
	if err != nil {
		return err
	}
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			return err
		}
		defer f.Close()
		out = f
	}
	return reporting.Write(out, *format, r)
}

func window(start, end string, last time.Duration) (telemetry.Window, error) {
	var w telemetry.Window
	if last != 0 {
		if start != "" || end != "" || last < 0 {
			return w, errors.New("choose --last or --start/--end")
		}
		w.End = time.Now().UTC().Truncate(time.Second)
		w.Start = w.End.Add(-last)
	} else {
		var err error
		w.Start, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return w, fmt.Errorf("start: %w", err)
		}
		w.End, err = time.Parse(time.RFC3339, end)
		if err != nil {
			return w, fmt.Errorf("end: %w", err)
		}
	}
	return w, w.Validate()
}

func encode(w io.Writer, v any) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(v)
}
