package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Prometheus   string            `json:"prometheus"`
	Tempo        string            `json:"tempo"`
	Namespace    string            `json:"namespace"`
	Queries      map[string]string `json:"queries"`
	TraceLimit   int               `json:"trace_limit"`
	CPUNearLimit float64           `json:"cpu_near_limit"`
	DBUser       string            `json:"db_user"`
}

func DefaultConfig() Config {
	return Config{Prometheus: "http://localhost:9090", Tempo: "http://localhost:3200", Namespace: "scalens", TraceLimit: 30, CPUNearLimit: .8, DBUser: "scalens"}
}

type Client struct {
	Config Config
	HTTP   *http.Client
	// Set after retries fail once, so a dead backend is not retried per query.
	unreachable bool
}

func New(c Config) *Client { return &Client{Config: c, HTTP: &http.Client{Timeout: 10 * time.Second}} }

func (c *Client) get(ctx context.Context, base, path string, params url.Values, out any) error {
	u, err := url.Parse(base)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("invalid backend URL")
	}
	u.Path = strings.TrimRight(u.Path, "/") + path
	u.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTP.Do(req)
	for attempt := 0; err != nil && attempt < 2 && !c.unreachable && ctx.Err() == nil; attempt++ {
		time.Sleep(time.Second)
		resp, err = c.HTTP.Do(req)
	}
	if err != nil {
		c.unreachable = true
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("%s returned HTTP %d", path, resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(out)
}

type promResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Value  []json.RawMessage   `json:"value"`
			Values [][]json.RawMessage `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func parseValue(pair []json.RawMessage) (float64, error) {
	if len(pair) != 2 {
		return 0, fmt.Errorf("invalid sample")
	}
	var text string
	if err := json.Unmarshal(pair[1], &text); err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(text, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("non-finite sample")
	}
	return v, nil
}

func (c *Client) metric(ctx context.Context, q, unit string, w Window) Measurement {
	m := Measurement{Unit: unit, Query: q}
	var data promResponse
	err := c.get(ctx, c.Config.Prometheus, "/api/v1/query", url.Values{"query": {q}, "time": {strconv.FormatInt(w.End.Unix(), 10)}}, &data)
	if err == nil && data.Status != "success" {
		err = fmt.Errorf("Prometheus: %s", data.Error)
	}
	if err == nil && len(data.Data.Result) != 1 {
		err = fmt.Errorf("expected one aggregate series, got %d", len(data.Data.Result))
	}
	if err == nil {
		var v float64
		v, err = parseValue(data.Data.Result[0].Value)
		if err == nil {
			m.Value = &v
			m.Samples = 1
		}
	}
	if err != nil {
		m.Error = err.Error()
	}
	return m
}

func (c *Client) cpu(ctx context.Context, q string, w Window) (Measurement, Measurement) {
	usage := Measurement{Query: q, Unit: "ratio"}
	near := Measurement{Query: q, Unit: "fraction_of_samples"}
	var data promResponse
	err := c.get(ctx, c.Config.Prometheus, "/api/v1/query_range", url.Values{
		"query": {q}, "start": {strconv.FormatInt(w.Start.Add(30*time.Second).Unix(), 10)},
		"end": {strconv.FormatInt(w.End.Add(-time.Second).Unix(), 10)}, "step": {"5"},
	}, &data)
	if err == nil && (data.Status != "success" || len(data.Data.Result) != 1) {
		err = fmt.Errorf("CPU series unavailable")
	}
	if err != nil {
		usage.Error = err.Error()
		near.Error = err.Error()
		return usage, near
	}
	var sum, high float64
	for _, pair := range data.Data.Result[0].Values {
		v, err := parseValue(pair)
		if err != nil {
			continue
		}
		sum += v
		if v >= c.Config.CPUNearLimit {
			high++
		}
		usage.Samples++
	}
	near.Samples = usage.Samples
	if usage.Samples < 3 {
		usage.Error = "fewer than three valid CPU samples"
		near.Error = usage.Error
		return usage, near
	}
	avg := sum / float64(usage.Samples)
	fraction := high / float64(usage.Samples)
	usage.Value = &avg
	near.Value = &fraction
	return usage, near
}

func (c *Client) Collect(ctx context.Context, t Target, w Window, slo float64) (Snapshot, error) {
	if err := w.Validate(); err != nil {
		return Snapshot{}, err
	}
	if t.Service == "" || t.Route == "" || t.Namespace == "" {
		return Snapshot{}, fmt.Errorf("service, route and namespace are required")
	}
	if c.Config.TraceLimit < 1 || c.Config.TraceLimit > 200 {
		return Snapshot{}, fmt.Errorf("trace_limit must be 1..200")
	}
	s := Snapshot{Schema: Schema, CapturedAt: time.Now().UTC(), Target: t, Window: w, Metrics: map[string]Measurement{}, SLOSeconds: slo, TraceLimit: c.Config.TraceLimit, Sampling: "Tempo bounded first-N search, split across time windows; not random or representative", Source: map[string]string{"prometheus": c.Config.Prometheus, "tempo": c.Config.Tempo}, RawTraces: map[string]json.RawMessage{}}
	sel := fmt.Sprintf(`service=%q,route=%q,method="POST",namespace=%q`, t.Service, t.Route, t.Namespace)
	rng := fmt.Sprintf("%ds", int(w.End.Sub(w.Start).Seconds()))
	count := fmt.Sprintf(`sum(increase(http_server_requests_total{%s}[%s]))`, sel, rng)
	success := fmt.Sprintf(`sum(increase(http_server_requests_total{%s,outcome="success"}[%s]))`, sel, rng)
	dbUser := c.Config.DBUser
	if dbUser == "" {
		dbUser = ".+"
	}
	queries := map[string]string{
		"requests":       count,
		"successes":      success,
		"error_rate":     fmt.Sprintf(`(%s - (%s or vector(0))) / %s`, count, success, count),
		"success_rps":    fmt.Sprintf(`(%s or vector(0)) / %g`, success, w.End.Sub(w.Start).Seconds()),
		"total_rps":      fmt.Sprintf(`%s / %g`, count, w.End.Sub(w.Start).Seconds()),
		"query_p95":      fmt.Sprintf(`histogram_quantile(0.95, sum by(le) (increase(db_client_query_duration_seconds_bucket{service=%q,namespace=%q,operation="OrderRepository.CountRecentOrders",outcome="success"}[%s])))`, t.Service, t.Namespace, rng),
		"acquire_p95":    fmt.Sprintf(`histogram_quantile(0.95, sum by(le) (increase(db_client_acquire_duration_seconds_bucket{service=%q,namespace=%q,outcome="success"}[%s])))`, t.Service, t.Namespace, rng),
		"db_server_mean": fmt.Sprintf(`max((sum by(queryid) (increase(pg_stat_statements_seconds_total{user=~%q}[%s])) / (sum by(queryid) (increase(pg_stat_statements_calls_total{user=~%q}[%s])) > 0)) and topk(1, sum by(queryid) (increase(pg_stat_statements_seconds_total{user=~%q}[%s]))))`, dbUser, rng, dbUser, rng, dbUser, rng),
		"scrape_health":  fmt.Sprintf(`min(avg_over_time(up{job="application",namespace=%q}[%s]))`, t.Namespace, rng),
	}
	for _, outcome := range []string{"success", "failure"} {
		for _, p := range []int{50, 95, 99} {
			queries[fmt.Sprintf("%s_p%d", outcome, p)] = fmt.Sprintf(`histogram_quantile(%g, sum by(le) (increase(http_server_request_duration_seconds_bucket{%s,outcome=%q}[%s])))`, float64(p)/100, sel, outcome, rng)
		}
	}
	if slo > 0 {
		queries["slo_bad_rate"] = fmt.Sprintf(`1 - ((sum(increase(http_server_request_duration_seconds_bucket{%s,outcome="success",le=%q}[%s])) or vector(0)) / %s)`, sel, strconv.FormatFloat(slo, 'f', -1, 64), rng, count)
		queries["slo_bucket"] = fmt.Sprintf(`count(http_server_request_duration_seconds_bucket{%s,outcome="success",le=%q})`, sel, strconv.FormatFloat(slo, 'f', -1, 64))
	}
	podSelector := fmt.Sprintf(`namespace=%q,container=%q,pod=~%q`, t.Namespace, t.Service, t.Service+"-.*")
	cpuQ := fmt.Sprintf(`max(rate(container_cpu_usage_seconds_total{%s}[30s]) / on(namespace,pod,container) group_left kube_pod_container_resource_limits{%s,resource="cpu",unit="core"})`, podSelector, podSelector)
	queries["cpu_ratio"] = cpuQ
	queries["cpu_throttled_ratio"] = fmt.Sprintf(`sum(increase(container_cpu_cfs_throttled_periods_total{%s}[%s])) / sum(increase(container_cpu_cfs_periods_total{%s}[%s]))`, podSelector, rng, podSelector, rng)
	keys := make([]string, 0, len(queries))
	for key := range queries {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		q := queries[key]
		if custom := c.Config.Queries[key]; custom != "" {
			q = strings.NewReplacer("$SERVICE", strconv.Quote(t.Service), "$ROUTE", strconv.Quote(t.Route), "$NAMESPACE", strconv.Quote(t.Namespace), "$RANGE", rng).Replace(custom)
		}
		if key == "cpu_ratio" {
			s.Metrics["cpu_ratio"], s.Metrics["cpu_near_limit_fraction"] = c.cpu(ctx, q, w)
		} else {
			unit := "ratio"
			if strings.Contains(key, "_p") || key == "db_server_mean" {
				unit = "seconds"
			}
			if strings.HasSuffix(key, "rps") {
				unit = "requests/second"
			}
			if key == "requests" || key == "successes" {
				unit = "requests"
			}
			s.Metrics[key] = c.metric(ctx, q, unit, w)
		}
	}
	if slo > 0 {
		if n, ok := s.Value("slo_bucket"); !ok || n < 1 {
			m := s.Metrics["slo_bad_rate"]
			m.Value = nil
			m.Error = "SLO bucket not present"
			s.Metrics["slo_bad_rate"] = m
		}
	}
	for key, m := range s.Metrics {
		if rate, ok := s.Value("error_rate"); ok && rate == 0 && strings.HasPrefix(key, "failure_") {
			continue
		}
		if m.Error != "" {
			s.Missing = append(s.Missing, key+": "+m.Error)
		}
	}
	c.collectTraces(ctx, &s)
	sort.Strings(s.Missing)
	return s, nil
}
