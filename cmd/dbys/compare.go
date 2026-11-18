package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/panjitamzil/scalens/internal/experiments"
)

func compareCommand(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	control := fs.String("control", "", "control run directory")
	candidate := fs.String("candidate", "", "candidate run directory")
	format := fs.String("format", "markdown", "json or markdown")
	flagged := fs.Bool("include-flagged", false, "include runs flagged for host contention or dropped iterations and label the outcome")
	if err := fs.Parse(args); err != nil {
		return err
	}
	a, err := experiments.ReadRuns(*control)
	if err != nil {
		return err
	}
	b, err := experiments.ReadRuns(*candidate)
	if err != nil {
		return err
	}
	result, err := experiments.Compare(a, b, *flagged)
	if err != nil {
		return err
	}
	if *format == "json" {
		return encode(out, result)
	}
	if *format != "markdown" {
		return fmt.Errorf("format must be json or markdown")
	}
	fmt.Fprintf(out, "# Remediation comparison\n\nStatus: **%s** (evidence: %s).\n\n| Measurement | Control | Candidate |\n|---|---:|---:|\n| Runs used | %d | %d |\n| Median successful p95 (ms) | %.3f | %.3f |\n| Median successful requests/s | %.3f | %.3f |\n| Median error rate | %.4f | %.4f |\n| Median SLO bad-event rate | %.4f | %.4f |\n\n", result.Status, result.Quality, result.Control.Runs, result.Candidate.Runs, result.Control.P95Median, result.Candidate.P95Median, result.Control.SuccessRPS, result.Candidate.SuccessRPS, result.Control.ErrorRate, result.Candidate.ErrorRate, result.Control.SLOBadRate, result.Candidate.SLOBadRate)
	for _, n := range result.Notes {
		fmt.Fprintf(out, "- %s\n", n)
	}
	for _, n := range result.Excluded {
		fmt.Fprintf(out, "- Excluded: %s\n", n)
	}
	return nil
}
