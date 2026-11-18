# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-16T07:05:56Z — 2026-11-16T07:07:56Z.

Status: **suspected** (suspected causes are not verified).

## Observations

- Error rate 0.517 exceeds configured 0.010
- SLO bad-event rate 0.783 exceeds 0.050
- Client-side database p95 is 0.876s but the costliest statement averages 0.0072s inside PostgreSQL; the time is spent outside query execution

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 4.2203 | seconds |
| cpu_near_limit_fraction | 0.76923 | fraction_of_samples |
| cpu_ratio | 0.86145 | ratio |
| cpu_throttled_ratio | 0.7361 | ratio |
| db_server_mean | 0.0071536 | seconds |
| error_rate | 0.51666 | ratio |
| failure_p50 | 7.5521 | seconds |
| failure_p95 | 9.9147 | seconds |
| failure_p99 | 10 | seconds |
| query_p95 | 0.87597 | seconds |
| requests | 2035 | requests |
| scrape_health | 0.78947 | ratio |
| slo_bad_rate | 0.78257 | ratio |
| slo_bucket | 2 | ratio |
| success_p50 | 1.0659 | seconds |
| success_p95 | 8.2787 | seconds |
| success_p99 | 9.6557 | seconds |
| success_rps | 8.1967 | requests/second |
| successes | 983.61 | requests |
| total_rps | 16.959 | requests/second |

## Suspect: cpu_saturation

Operations: []

- Evidence: `cpu_ratio`
- Evidence: `cpu_near_limit_fraction`
- Evidence: `cpu_throttled_ratio`
- Limitation: Container CPU evidence does not identify a CPU-expensive function
- Limitation: Host contention must be checked independently
- Experiment: Compare one and two replicas at the same offered load with host headroom
- Experiment: Inspect CPU profiles if local computation remains unclear

Traces analyzed: 29; incomplete: 0; candidates: 30.

Sampling: Tempo bounded first-N search, split across time windows; not random or representative.

## Missing evidence and warnings

- application scrape coverage below 95% or unknown
- Scrapes were missed during the window; counters are under-sampled and an overloaded target can cause this
- Sampled traces are selected evidence, not a population latency estimate

## Metric provenance

### acquire_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_acquire_duration_seconds_bucket{service="order-service",namespace="scalens",outcome="success"}[120s])))
```

### cpu_near_limit_fraction

```promql
max(rate(container_cpu_usage_seconds_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[30s]) / on(namespace,pod,container) group_left kube_pod_container_resource_limits{namespace="scalens",container="order-service",pod=~"order-service-.*",resource="cpu",unit="core"})
```

### cpu_ratio

```promql
max(rate(container_cpu_usage_seconds_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[30s]) / on(namespace,pod,container) group_left kube_pod_container_resource_limits{namespace="scalens",container="order-service",pod=~"order-service-.*",resource="cpu",unit="core"})
```

### cpu_throttled_ratio

```promql
sum(increase(container_cpu_cfs_throttled_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[120s])) / sum(increase(container_cpu_cfs_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[120s]))
```

### db_server_mean

```promql
max((sum by(queryid) (increase(pg_stat_statements_seconds_total[120s])) / (sum by(queryid) (increase(pg_stat_statements_calls_total[120s])) > 0)) and topk(1, sum by(queryid) (increase(pg_stat_statements_seconds_total[120s]))))
```

### error_rate

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[120s])) - (sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[120s])) or vector(0))) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[120s]))
```

### failure_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[120s])))
```

### failure_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[120s])))
```

### failure_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[120s])))
```

### query_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_query_duration_seconds_bucket{service="order-service",namespace="scalens",operation="OrderRepository.CountRecentOrders",outcome="success"}[120s])))
```

### requests

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[120s]))
```

### scrape_health

```promql
min(avg_over_time(up{job="application",namespace="scalens"}[120s]))
```

### slo_bad_rate

```promql
1 - ((sum(increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"}[120s])) or vector(0)) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[120s])))
```

### slo_bucket

```promql
count(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"})
```

### success_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[120s])))
```

### success_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[120s])))
```

### success_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[120s])))
```

### success_rps

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[120s])) or vector(0)) / 120
```

### successes

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[120s]))
```

### total_rps

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[120s])) / 120
```
