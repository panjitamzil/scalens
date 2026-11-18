# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-16T15:46:33Z — 2026-11-16T15:48:42Z.

Status: **suspected** (suspected causes are not verified).

## Observations

- Error rate 0.998 exceeds configured 0.010
- SLO bad-event rate 1.000 exceeds 0.050

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 9.7419 | seconds |
| cpu_near_limit_fraction | 0 | fraction_of_samples |
| cpu_ratio | 0.13983 | ratio |
| cpu_throttled_ratio | 0.0017301 | ratio |
| db_server_mean | 1.3574 | seconds |
| error_rate | 0.99817 | ratio |
| failure_p50 | 7.5 | seconds |
| failure_p95 | 9.75 | seconds |
| failure_p99 | 9.95 | seconds |
| query_p95 | 3 | seconds |
| requests | 2256.9 | requests |
| scrape_health | 1 | ratio |
| slo_bad_rate | 0.99954 | ratio |
| slo_bucket | 1 | ratio |
| success_p50 | 6.6667 | seconds |
| success_p95 | 9.6667 | seconds |
| success_p99 | 9.9333 | seconds |
| success_rps | 0.031999 | requests/second |
| successes | 4.1279 | requests |
| total_rps | 17.496 | requests/second |

## Suspect: slow_database_operation

Operations: [OrderRepository.CountRecentOrders]

- Evidence: `query_p95`
- Evidence: `db_server_mean`
- Evidence: `trace:c64fa5eb21d11681520b56fed69ef64`
- Evidence: `trace:7660dc1349c8510156b8b4b1df636fc`
- Evidence: `trace:12b405427c73cd4f0de2f26e12939e4d`
- Evidence: `trace:cc517b138ab730715d2e75ae3b9f0c6`
- Evidence: `trace:fffa7706988f0c70114c6d89ffd5310`
- Evidence: `trace:f730e2090d521d20576fdeda8e82bf9`
- Evidence: `trace:e09da4759c099b01719eacbd119a00b`
- Evidence: `trace:855c15038c610e903f909f6b2d134aa`
- Evidence: `trace:bec7cc5e4b51b8505de7a344ef3fc42`
- Evidence: `trace:fcc96792d664f650b5fa6650cbb2847`
- Evidence: `trace:32971f8b4c918ca160cf0d04d952757`
- Evidence: `trace:6f8e83ea6a0d9e093743582dc59ec6`
- Evidence: `trace:143cda2706dc1a710d4a82b6c43eda7f`
- Evidence: `trace:132c34a3634ea0770146a5bf3ea6b63c`
- Evidence: `trace:11998c43d7e3b961797cd451b55401d`
- Evidence: `trace:d6debb4f976ad020474d2c963d65a5a`
- Evidence: `trace:8a97716876fa611a05bcc440a372250b`
- Evidence: `trace:eb7fa739810a712e04699ca71c9a386d`
- Evidence: `trace:c82e7daae9f0201c08b34496e881aa47`
- Evidence: `trace:b61ba12b90d9d6370c2ca866fcfb1ed6`
- Evidence: `trace:4c5c70a72f5f90c214d0b51da8630ebd`
- Evidence: `trace:1651f039f5143d98051bece514ee62d7`
- Evidence: `trace:f797a33c03ca9b106c107faf5c98dfb`
- Evidence: `trace:79c79e09405af920c900fd3a54d1f6e`
- Limitation: Statement time includes execution, lock and I/O waiting
- Limitation: Neither a missing index nor PostgreSQL saturation is proven by slow statements
- Limitation: Connection acquisition is slower than query execution; pool waiting may also contribute
- Experiment: Inspect EXPLAIN (ANALYZE, BUFFERS) and pg_stat_statements for the operation
- Experiment: Compare the relevant index change with application scale-out

Traces analyzed: 30; incomplete: 0; candidates: 30.

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
