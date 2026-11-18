package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"time"
)

type attribute struct {
	Key   string `json:"key"`
	Value struct {
		String string `json:"stringValue"`
		Int    string `json:"intValue"`
	} `json:"value"`
}

func attr(attrs []attribute, key string) string {
	for _, a := range attrs {
		if a.Key == key {
			return a.Value.String
		}
	}
	return ""
}

type Span struct {
	ID         string      `json:"spanId"`
	Parent     string      `json:"parentSpanId"`
	Name       string      `json:"name"`
	Start      string      `json:"startTimeUnixNano"`
	End        string      `json:"endTimeUnixNano"`
	Attributes []attribute `json:"attributes"`
	Service    string      `json:"-"`
}

type scope struct {
	Spans []Span `json:"spans"`
}
type batch struct {
	Resource struct {
		Attributes []attribute `json:"attributes"`
	} `json:"resource"`
	Scopes []scope `json:"scopeSpans"`
	Legacy []scope `json:"instrumentationLibrarySpans"`
}

type traceData struct {
	Batches   []batch `json:"batches"`
	Resources []batch `json:"resourceSpans"`
}

type interval struct{ start, end int64 }

func union(intervals []interval) int64 {
	if len(intervals) == 0 {
		return 0
	}
	sort.Slice(intervals, func(i, j int) bool { return intervals[i].start < intervals[j].start })
	cur := intervals[0]
	var total int64
	for _, next := range intervals[1:] {
		if next.start <= cur.end {
			if next.end > cur.end {
				cur.end = next.end
			}
			continue
		}
		total += cur.end - cur.start
		cur = next
	}
	return total + cur.end - cur.start
}

func bounds(s Span) (interval, bool) {
	start, e1 := strconv.ParseInt(s.Start, 10, 64)
	end, e2 := strconv.ParseInt(s.End, 10, 64)
	return interval{start, end}, e1 == nil && e2 == nil && end >= start
}

func clip(v, parent interval) (interval, bool) {
	if v.start < parent.start {
		v.start = parent.start
	}
	if v.end > parent.end {
		v.end = parent.end
	}
	return v, v.end > v.start
}

func AnalyzeTrace(id string, spans []Span, target Target) TraceSummary {
	out := TraceSummary{ID: id}
	ids := map[string]Span{}
	times := map[string]interval{}
	var root Span
	var rootTime interval
	for _, s := range spans {
		if _, exists := ids[s.ID]; exists {
			out.Incomplete = true
		}
		ids[s.ID] = s
		v, ok := bounds(s)
		if !ok {
			out.Incomplete = true
			continue
		}
		times[s.ID] = v
		if s.Service == target.Service && attr(s.Attributes, "http.route") == target.Route && (root.ID == "" || v.end-v.start > rootTime.end-rootTime.start) {
			root = s
			rootTime = v
		}
	}
	if root.ID == "" {
		out.Incomplete = true
		return out
	}
	var query, acquire, children []interval
	ops := map[string][]interval{}
	for _, s := range spans {
		v, ok := times[s.ID]
		if !ok {
			continue
		}
		if s.Parent != "" {
			if _, exists := ids[s.Parent]; !exists {
				out.Incomplete = true
			}
		}
		if s.Service != target.Service {
			continue
		}
		clipped, ok := clip(v, rootTime)
		if !ok {
			continue
		}
		if s.Parent == root.ID {
			children = append(children, clipped)
		}
		switch attr(s.Attributes, "dbys.phase") {
		case "query_execute":
			query = append(query, clipped)
			ops[s.Name] = append(ops[s.Name], clipped)
		case "connection_acquire":
			acquire = append(acquire, clipped)
		}
	}
	out.RootDuration = float64(rootTime.end-rootTime.start) / 1e9
	out.QuerySeconds = float64(union(query)) / 1e9
	out.AcquireSeconds = float64(union(acquire)) / 1e9
	if !out.Incomplete {
		out.LocalSeconds = float64(rootTime.end-rootTime.start-union(children)) / 1e9
	}
	out.Operations = map[string]float64{}
	for name, spans := range ops {
		out.Operations[name] = float64(union(spans)) / 1e9
	}
	return out
}

func (c *Client) collectTraces(ctx context.Context, s *Snapshot) {
	q := fmt.Sprintf(`{ resource.service.name = %q && span.http.route = %q }`, s.Target.Service, s.Target.Route)
	s.TraceQuery = q
	ids := map[string]bool{}
	for part := 0; part < 3; part++ {
		start := s.Window.Start.Add(time.Duration(part) * s.Window.End.Sub(s.Window.Start) / 3)
		end := s.Window.Start.Add(time.Duration(part+1) * s.Window.End.Sub(s.Window.Start) / 3)
		var found struct {
			Traces []struct {
				ID string `json:"traceID"`
			} `json:"traces"`
		}
		err := c.get(ctx, c.Config.Tempo, "/api/search", url.Values{"q": {q}, "start": {strconv.FormatInt(start.Unix(), 10)}, "end": {strconv.FormatInt(end.Unix(), 10)}, "limit": {strconv.Itoa((c.Config.TraceLimit + 2) / 3)}}, &found)
		if err != nil {
			s.Missing = append(s.Missing, "traces: "+err.Error())
			return
		}
		s.TraceCandidates += len(found.Traces)
		for _, hit := range found.Traces {
			if ids[hit.ID] || len(ids) >= c.Config.TraceLimit {
				continue
			}
			ids[hit.ID] = true
			var raw json.RawMessage
			if err := c.get(ctx, c.Config.Tempo, "/api/traces/"+url.PathEscape(hit.ID), nil, &raw); err != nil {
				s.Missing = append(s.Missing, "trace "+hit.ID+": "+err.Error())
				continue
			}
			s.RawTraces[hit.ID] = raw
			var data traceData
			if err := json.Unmarshal(raw, &data); err != nil {
				s.Missing = append(s.Missing, "trace decode: "+err.Error())
				continue
			}
			var spans []Span
			for _, b := range append(data.Batches, data.Resources...) {
				service := attr(b.Resource.Attributes, "service.name")
				for _, scope := range append(b.Scopes, b.Legacy...) {
					for _, span := range scope.Spans {
						span.Service = service
						spans = append(spans, span)
					}
				}
			}
			summary := AnalyzeTrace(hit.ID, spans, s.Target)
			if summary.Incomplete {
				s.TraceIncomplete++
			}
			s.Traces = append(s.Traces, summary)
		}
	}
	if len(s.Traces) == 0 {
		s.Missing = append(s.Missing, "no traces found; search selection and export sampling may miss requests")
	}
}
