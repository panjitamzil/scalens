# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-16T15:49:41Z — 2026-11-16T15:51:50Z.

Status: **suspected** (suspected causes are not verified).

## Observations

- Error rate 0.997 exceeds configured 0.010
- SLO bad-event rate 1.000 exceeds 0.050

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 9.7402 | seconds |
| cpu_near_limit_fraction | 0 | fraction_of_samples |
| cpu_ratio | 0.15297 | ratio |
| cpu_throttled_ratio | 0.016575 | ratio |
| db_server_mean | 1.6145 | seconds |
| error_rate | 0.9973 | ratio |
| failure_p50 | 7.5 | seconds |
| failure_p95 | 9.75 | seconds |
| failure_p99 | 9.95 | seconds |
| query_p95 | 3.8542 | seconds |
| requests | 2293.1 | requests |
| scrape_health | 1 | ratio |
| slo_bad_rate | 1 | ratio |
| slo_bucket | 1 | ratio |
| success_p50 | 6.25 | seconds |
| success_p95 | 9.625 | seconds |
| success_p99 | 9.925 | seconds |
| success_rps | 0.048 | requests/second |
| successes | 6.192 | requests |
| total_rps | 17.776 | requests/second |

## Suspect: slow_database_operation

Operations: [OrderRepository.CountRecentOrders]

- Evidence: `query_p95`
- Evidence: `db_server_mean`
- Evidence: `trace:194c1fac7fafaaa903ae2654542058c5`
- Evidence: `trace:2a14ea3f81498b180d4bf6816133d1b1`
- Evidence: `trace:74f4eec6cd8b17402b090fd23836361`
- Evidence: `trace:5b703a1ebd340550c9c8e7b7d967566`
- Evidence: `trace:a77c4a59b19f4090398002b7b5800a4`
- Evidence: `trace:11aab061858d5533049087913066b57a`
- Evidence: `trace:256cb5d1d7fa62840c357ce0deda5314`
- Evidence: `trace:1e1a09c988b5f4a50d49e3c39df68eba`
- Evidence: `trace:8f049325dde7859094c764c0e0f0991`
- Evidence: `trace:e16acdd689b4145051dc2af0b00b58d`
- Evidence: `trace:dcc7847bd3dc96c1612de5dd3e4633c`
- Evidence: `trace:5429e03e7bf69a5d17abcbed4ad6c1b5`
- Evidence: `trace:568ff95f3d655b1a1778388cb45e8a33`
- Evidence: `trace:9ceb26df856ddfc41492cf8dd035e6a2`
- Evidence: `trace:cc8e89810c8b67fd02355d196cdea175`
- Evidence: `trace:e2616288ce523f5209810d42e36d59b4`
- Limitation: Statement time includes execution, lock and I/O waiting
- Limitation: Neither a missing index nor PostgreSQL saturation is proven by slow statements
- Limitation: Connection acquisition is slower than query execution; pool waiting may also contribute
- Experiment: Inspect EXPLAIN (ANALYZE, BUFFERS) and pg_stat_statements for the operation
- Experiment: Compare the relevant index change with application scale-out

Traces analyzed: 27; incomplete: 0; candidates: 30.

Sampling: Tempo bounded first-N search, split across time windows; not random or representative.

## Missing evidence and warnings

- Sampled traces are selected evidence, not a population latency estimate

## Metric provenance

### acquire_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_acquire_duration_seconds_bucket{service="order-service",namespace="scalens",outcome="success"}[129s])))
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
sum(increase(container_cpu_cfs_throttled_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[129s])) / sum(increase(container_cpu_cfs_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[129s]))
```

### db_server_mean

```promql
max((sum by(queryid) (increase(pg_stat_statements_seconds_total[129s])) / (sum by(queryid) (increase(pg_stat_statements_calls_total[129s])) > 0)) and topk(1, sum by(queryid) (increase(pg_stat_statements_seconds_total[129s]))))
```

### error_rate

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[129s])) - (sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[129s])) or vector(0))) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[129s]))
```

### failure_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[129s])))
```

### failure_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[129s])))
```

### failure_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[129s])))
```

### query_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_query_duration_seconds_bucket{service="order-service",namespace="scalens",operation="OrderRepository.CountRecentOrders",outcome="success"}[129s])))
```

### requests

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[129s]))
```

### scrape_health

```promql
min(avg_over_time(up{job="application",namespace="scalens"}[129s]))
```

### slo_bad_rate

```promql
1 - ((sum(increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"}[129s])) or vector(0)) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[129s])))
```

### slo_bucket

```promql
count(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"})
```

### success_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[129s])))
```

### success_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[129s])))
```

### success_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[129s])))
```

### success_rps

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[129s])) or vector(0)) / 129
```

### successes

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[129s]))
```

### total_rps

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[129s])) / 129
```
