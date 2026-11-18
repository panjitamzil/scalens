# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-17T10:03:04Z — 2026-11-17T10:05:08Z.

Status: **suspected** (suspected causes are not verified).

## Observations

- Error rate 0.424 exceeds configured 0.010
- SLO bad-event rate 0.942 exceeds 0.050
- Client-side database p95 is 0.900s but the costliest statement averages 0.0002s inside PostgreSQL; the time is spent outside query execution

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 4.5386 | seconds |
| cpu_near_limit_fraction | 0.75 | fraction_of_samples |
| cpu_ratio | 0.92692 | ratio |
| cpu_throttled_ratio | 0.99908 | ratio |
| db_server_mean | 0.00021775 | seconds |
| error_rate | 0.42369 | ratio |
| failure_p50 | 7.5143 | seconds |
| failure_p95 | 9.7771 | seconds |
| failure_p99 | 9.9783 | seconds |
| query_p95 | 0.90021 | seconds |
| requests | 2146.2 | requests |
| scrape_health | 1 | ratio |
| slo_bad_rate | 0.94222 | ratio |
| slo_bucket | 1 | ratio |
| success_p50 | 2.0185 | seconds |
| success_p95 | 8.5473 | seconds |
| success_p99 | 9.7095 | seconds |
| success_rps | 9.975 | requests/second |
| successes | 1236.9 | requests |
| total_rps | 17.308 | requests/second |

## Suspect: cpu_saturation

Operations: []

- Evidence: `cpu_ratio`
- Evidence: `cpu_near_limit_fraction`
- Evidence: `cpu_throttled_ratio`
- Limitation: Container CPU evidence does not identify a CPU-expensive function
- Limitation: Host contention must be checked independently
- Experiment: Compare one and two replicas at the same offered load with host headroom
- Experiment: Inspect CPU profiles if local computation remains unclear

Traces analyzed: 26; incomplete: 0; candidates: 30.

Sampling: Tempo bounded first-N search, split across time windows; not random or representative.

## Missing evidence and warnings

- Sampled traces are selected evidence, not a population latency estimate

## Metric provenance

### acquire_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_acquire_duration_seconds_bucket{service="order-service",namespace="scalens",outcome="success"}[124s])))
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
sum(increase(container_cpu_cfs_throttled_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[124s])) / sum(increase(container_cpu_cfs_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[124s]))
```

### db_server_mean

```promql
max((sum by(queryid) (increase(pg_stat_statements_seconds_total{user=~"scalens"}[124s])) / (sum by(queryid) (increase(pg_stat_statements_calls_total{user=~"scalens"}[124s])) > 0)) and topk(1, sum by(queryid) (increase(pg_stat_statements_seconds_total{user=~"scalens"}[124s]))))
```

### error_rate

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[124s])) - (sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[124s])) or vector(0))) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[124s]))
```

### failure_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[124s])))
```

### failure_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[124s])))
```

### failure_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[124s])))
```

### query_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_query_duration_seconds_bucket{service="order-service",namespace="scalens",operation="OrderRepository.CountRecentOrders",outcome="success"}[124s])))
```

### requests

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[124s]))
```

### scrape_health

```promql
min(avg_over_time(up{job="application",namespace="scalens"}[124s]))
```

### slo_bad_rate

```promql
1 - ((sum(increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"}[124s])) or vector(0)) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[124s])))
```

### slo_bucket

```promql
count(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"})
```

### success_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[124s])))
```

### success_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[124s])))
```

### success_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[124s])))
```

### success_rps

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[124s])) or vector(0)) / 124
```

### successes

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[124s]))
```

### total_rps

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[124s])) / 124
```
