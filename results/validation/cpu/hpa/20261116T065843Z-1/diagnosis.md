# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-16T06:59:56Z — 2026-11-16T07:01:58Z.

Status: **insufficient_evidence** (suspected causes are not verified).

## Observations

- Error rate 0.190 exceeds configured 0.010
- SLO bad-event rate 0.457 exceeds 0.050
- Client-side database p95 is 0.470s but the costliest statement averages 0.0019s inside PostgreSQL; the time is spent outside query execution

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 2.4353 | seconds |
| cpu_near_limit_fraction | 0.21053 | fraction_of_samples |
| cpu_ratio | 0.7259 | ratio |
| cpu_throttled_ratio | 0.56886 | ratio |
| db_server_mean | 0.0019097 | seconds |
| error_rate | 0.1905 | ratio |
| failure_p50 | 7.3402 | seconds |
| failure_p95 | 9.8512 | seconds |
| failure_p99 | 10 | seconds |
| query_p95 | 0.47028 | seconds |
| requests | 2143.6 | requests |
| scrape_health | 1 | ratio |
| slo_bad_rate | 0.45692 | ratio |
| slo_bucket | 2 | ratio |
| success_p50 | 0.091892 | seconds |
| success_p95 | 8.0232 | seconds |
| success_p99 | 9.6234 | seconds |
| success_rps | 14.223 | requests/second |
| successes | 1735.3 | requests |
| total_rps | 17.571 | requests/second |

Traces analyzed: 30; incomplete: 0; candidates: 30.

Sampling: Tempo bounded first-N search, split across time windows; not random or representative.

## Missing evidence and warnings

- symptoms observed, but neither CPU nor database rule has enough supporting evidence
- Sampled traces are selected evidence, not a population latency estimate

## Metric provenance

### acquire_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_acquire_duration_seconds_bucket{service="order-service",namespace="scalens",outcome="success"}[122s])))
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
sum(increase(container_cpu_cfs_throttled_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[122s])) / sum(increase(container_cpu_cfs_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[122s]))
```

### db_server_mean

```promql
max((sum by(queryid) (increase(pg_stat_statements_seconds_total[122s])) / (sum by(queryid) (increase(pg_stat_statements_calls_total[122s])) > 0)) and topk(1, sum by(queryid) (increase(pg_stat_statements_seconds_total[122s]))))
```

### error_rate

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[122s])) - (sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[122s])) or vector(0))) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[122s]))
```

### failure_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[122s])))
```

### failure_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[122s])))
```

### failure_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[122s])))
```

### query_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_query_duration_seconds_bucket{service="order-service",namespace="scalens",operation="OrderRepository.CountRecentOrders",outcome="success"}[122s])))
```

### requests

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[122s]))
```

### scrape_health

```promql
min(avg_over_time(up{job="application",namespace="scalens"}[122s]))
```

### slo_bad_rate

```promql
1 - ((sum(increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"}[122s])) or vector(0)) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[122s])))
```

### slo_bucket

```promql
count(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"})
```

### success_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[122s])))
```

### success_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[122s])))
```

### success_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[122s])))
```

### success_rps

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[122s])) or vector(0)) / 122
```

### successes

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[122s]))
```

### total_rps

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[122s])) / 122
```
