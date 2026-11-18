package telemetry

import (
	"testing"
)

func TestUnionOverlappingChildren(t *testing.T) {
	got := union([]interval{{0, 10}, {2, 8}, {8, 15}, {20, 30}})
	if got != 25 {
		t.Fatalf("union=%d, want 25", got)
	}
}

func TestTraceUsesUnionAndFlagsMissingParent(t *testing.T) {
	route := attribute{Key: "http.route"}
	route.Value.String = "/orders"
	phase := attribute{Key: "dbys.phase"}
	phase.Value.String = "query_execute"
	spans := []Span{
		{ID: "root", Service: "orders", Start: "0", End: "1000000000", Attributes: []attribute{route}},
		{ID: "a", Parent: "root", Service: "orders", Name: "query-a", Start: "100000000", End: "700000000", Attributes: []attribute{phase}},
		{ID: "b", Parent: "root", Service: "orders", Name: "query-b", Start: "400000000", End: "900000000", Attributes: []attribute{phase}},
	}
	s := AnalyzeTrace("trace", spans, Target{Service: "orders", Route: "/orders"})
	if s.QuerySeconds != .8 || s.LocalSeconds != .2 || s.Incomplete || s.Operations["query-a"] != .6 || s.Operations["query-b"] != .5 {
		t.Fatalf("overlap counted incorrectly: %+v", s)
	}
	spans[1].Parent = "missing"
	s = AnalyzeTrace("trace", spans, Target{Service: "orders", Route: "/orders"})
	if !s.Incomplete || s.LocalSeconds != 0 {
		t.Fatalf("incomplete trace treated as complete: %+v", s)
	}
}
