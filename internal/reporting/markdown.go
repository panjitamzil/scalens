package reporting

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/panjitamzil/scalens/internal/diagnosis"
)

func Write(w io.Writer, format string, r diagnosis.Report) error {
	if format == "json" {
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		return e.Encode(r)
	}
	if format != "markdown" {
		return fmt.Errorf("format must be json or markdown")
	}
	fmt.Fprintf(w, "# Diagnosis: %s\n\nTarget: `%s %s` in `%s`.\n\nWindow: %s — %s.\n\nStatus: **%s** (suspected causes are not verified).\n\n", r.Snapshot.Target.Service, r.Snapshot.Target.Service, r.Snapshot.Target.Route, r.Snapshot.Target.Namespace, r.Snapshot.Window.Start.Format("2006-01-02T15:04:05Z07:00"), r.Snapshot.Window.End.Format("2006-01-02T15:04:05Z07:00"), r.Status)
	fmt.Fprint(w, "## Observations\n\n")
	for _, v := range r.Observations {
		fmt.Fprintf(w, "- %s\n", v)
	}
	fmt.Fprintln(w, "\n## Measurements\n\n| Signal | Value | Unit |\n|---|---:|---|")
	keys := make([]string, 0, len(r.Snapshot.Metrics))
	for k := range r.Snapshot.Metrics {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		m := r.Snapshot.Metrics[k]
		v := "unavailable"
		if m.Value != nil {
			v = fmt.Sprintf("%.5g", *m.Value)
		}
		fmt.Fprintf(w, "| %s | %s | %s |\n", k, v, m.Unit)
	}
	for _, h := range r.Suspects {
		fmt.Fprintf(w, "\n## Suspect: %s\n\nOperations: %v\n\n", h.Category, h.Operations)
		for _, v := range h.Evidence {
			fmt.Fprintf(w, "- Evidence: `%s`\n", v)
		}
		for _, v := range h.Limitations {
			fmt.Fprintf(w, "- Limitation: %s\n", v)
		}
		for _, v := range h.Experiments {
			fmt.Fprintf(w, "- Experiment: %s\n", v)
		}
	}
	fmt.Fprintf(w, "\nTraces analyzed: %d; incomplete: %d; candidates: %d.\n\nSampling: %s.\n", len(r.Snapshot.Traces), r.Snapshot.TraceIncomplete, r.Snapshot.TraceCandidates, r.Snapshot.Sampling)
	fmt.Fprint(w, "\n## Missing evidence and warnings\n\n")
	for _, v := range append(r.Missing, r.Warnings...) {
		fmt.Fprintf(w, "- %s\n", v)
	}
	fmt.Fprintln(w, "\n## Metric provenance")
	for _, k := range keys {
		fmt.Fprintf(w, "\n### %s\n\n```promql\n%s\n```\n", k, r.Snapshot.Metrics[k].Query)
	}
	return nil
}
