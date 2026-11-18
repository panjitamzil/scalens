# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-17T09:54:37Z — 2026-11-17T09:56:38Z.

Status: **insufficient_evidence** (suspected causes are not verified).

## Observations

- SLO bad-event rate 0.084 exceeds 0.050
- Client-side database p95 is 0.054s but the costliest statement averages 0.0001s inside PostgreSQL; the time is spent outside query execution

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 0.047635 | seconds |
| cpu_near_limit_fraction | 0 | fraction_of_samples |
| cpu_ratio | 0.50506 | ratio |
| cpu_throttled_ratio | 0.099351 | ratio |
| db_server_mean | 0.00012861 | seconds |
| error_rate | 0 | ratio |
| failure_p50 | unavailable | seconds |
| failure_p95 | unavailable | seconds |
| failure_p99 | unavailable | seconds |
| query_p95 | 0.053822 | seconds |
| requests | 2371.7 | requests |
| scrape_health | 1 | ratio |
| slo_bad_rate | 0.083843 | ratio |
| slo_bucket | 2 | ratio |
| success_p50 | 0.039937 | seconds |
| success_p95 | 1.2144 | seconds |
| success_p99 | 2.4008 | seconds |
| success_rps | 19.601 | requests/second |
| successes | 2371.7 | requests |
| total_rps | 19.601 | requests/second |

Traces analyzed: 30; incomplete: 0; candidates: 30.

Sampling: Tempo bounded first-N search, split across time windows; not random or representative.

## Missing evidence and warnings

- symptoms observed, but neither CPU nor database rule has enough supporting evidence
- Sampled traces are selected evidence, not a population latency estimate

## Metric provenance

### acquire_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_acquire_duration_seconds_bucket{service="order-service",namespace="scalens",outcome="success"}[121s])))
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
sum(increase(container_cpu_cfs_throttled_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[121s])) / sum(increase(container_cpu_cfs_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[121s]))
```

### db_server_mean

```promql
max((sum by(queryid) (increase(pg_stat_statements_seconds_total{user=~"scalens"}[121s])) / (sum by(queryid) (increase(pg_stat_statements_calls_total{user=~"scalens"}[121s])) > 0)) and topk(1, sum by(queryid) (increase(pg_stat_statements_seconds_total{user=~"scalens"}[121s]))))
```

### error_rate

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[121s])) - (sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[121s])) or vector(0))) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[121s]))
```

### failure_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[121s])))
```

### failure_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[121s])))
```

### failure_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[121s])))
```

### query_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_query_duration_seconds_bucket{service="order-service",namespace="scalens",operation="OrderRepository.CountRecentOrders",outcome="success"}[121s])))
```

### requests

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[121s]))
```

### scrape_health

```promql
min(avg_over_time(up{job="application",namespace="scalens"}[121s]))
```

### slo_bad_rate

```promql
1 - ((sum(increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"}[121s])) or vector(0)) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[121s])))
```

### slo_bucket

```promql
count(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"})
```

### success_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[121s])))
```

### success_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[121s])))
```

### success_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[121s])))
```

### success_rps

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[121s])) or vector(0)) / 121
```

### successes

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[121s]))
```

### total_rps

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[121s])) / 121
```
