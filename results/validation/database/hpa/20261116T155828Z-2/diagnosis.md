# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-16T15:59:10Z — 2026-11-16T16:01:18Z.

Status: **suspected** (suspected causes are not verified).

## Observations

- Error rate 1.000 exceeds configured 0.010
- SLO bad-event rate 1.000 exceeds 0.050

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 9.747 | seconds |
| cpu_near_limit_fraction | 0 | fraction_of_samples |
| cpu_ratio | 0.1364 | ratio |
| cpu_throttled_ratio | 0.0017528 | ratio |
| db_server_mean | 1.7508 | seconds |
| error_rate | 0.99953 | ratio |
| failure_p50 | 7.5 | seconds |
| failure_p95 | 9.75 | seconds |
| failure_p99 | 9.95 | seconds |
| query_p95 | 4.3571 | seconds |
| requests | 2289.1 | requests |
| scrape_health | 1 | ratio |
| slo_bad_rate | 1 | ratio |
| slo_bucket | 1 | ratio |
| success_p50 | 1.75 | seconds |
| success_p95 | 2.425 | seconds |
| success_p99 | 2.485 | seconds |
| success_rps | 0.0083336 | requests/second |
| successes | 1.0667 | requests |
| total_rps | 17.884 | requests/second |

## Suspect: slow_database_operation

Operations: [OrderRepository.CountRecentOrders]

- Evidence: `query_p95`
- Evidence: `db_server_mean`
- Evidence: `trace:340f6f34cd235aab13eb4c0067f435e4`
- Evidence: `trace:6243ac97704925170791da309eabd5a2`
- Evidence: `trace:8276c88e05f540e504d598c4ea1ec9c7`
- Evidence: `trace:3f43680b3761421310737d0b88bbe02a`
- Evidence: `trace:1e5fc4fad6814da200fef898800f994c`
- Evidence: `trace:1a057a9c5d704e4c10401fc9f5166e3e`
- Evidence: `trace:2b172cc199766c4e043a76aacb6140d5`
- Evidence: `trace:37b2d0d9e6451cc040c473985a0482b`
- Evidence: `trace:110bd96b9230d61e07c5e88226c483be`
- Evidence: `trace:c259ade3507230f0f55293b1f5d49f9`
- Evidence: `trace:15da43769fecefc602928f9420e307e1`
- Evidence: `trace:cc7482426dde7f3192b06bdcd54b91d`
- Evidence: `trace:cbff423ef3ebb6d09494f8eb884b8ad`
- Evidence: `trace:be42b35525672d10301037848ed76c7`
- Evidence: `trace:10e806f508d0a6770ca67ab00a1b4785`
- Evidence: `trace:7cbcb2d12a3dae508d9f9e647bc47eb`
- Evidence: `trace:1380a2863638ac5204b738a2b23d697a`
- Evidence: `trace:409b163efc497ab21527521fda2da2e5`
- Evidence: `trace:e9ff93ddc62b1300105bdce3dfb8cab`
- Evidence: `trace:4dd01152045069ae03986f557c873a9f`
- Evidence: `trace:4a4f6821b47c0df415175f2cf9efd627`
- Evidence: `trace:e43ef55b1859e980a6361eb0eb69880`
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
histogram_quantile(0.95, sum by(le) (increase(db_client_acquire_duration_seconds_bucket{service="order-service",namespace="scalens",outcome="success"}[128s])))
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
sum(increase(container_cpu_cfs_throttled_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[128s])) / sum(increase(container_cpu_cfs_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[128s]))
```

### db_server_mean

```promql
max((sum by(queryid) (increase(pg_stat_statements_seconds_total[128s])) / (sum by(queryid) (increase(pg_stat_statements_calls_total[128s])) > 0)) and topk(1, sum by(queryid) (increase(pg_stat_statements_seconds_total[128s]))))
```

### error_rate

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[128s])) - (sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[128s])) or vector(0))) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[128s]))
```

### failure_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[128s])))
```

### failure_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[128s])))
```

### failure_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[128s])))
```

### query_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_query_duration_seconds_bucket{service="order-service",namespace="scalens",operation="OrderRepository.CountRecentOrders",outcome="success"}[128s])))
```

### requests

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[128s]))
```

### scrape_health

```promql
min(avg_over_time(up{job="application",namespace="scalens"}[128s]))
```

### slo_bad_rate

```promql
1 - ((sum(increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"}[128s])) or vector(0)) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[128s])))
```

### slo_bucket

```promql
count(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"})
```

### success_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[128s])))
```

### success_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[128s])))
```

### success_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[128s])))
```

### success_rps

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[128s])) or vector(0)) / 128
```

### successes

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[128s]))
```

### total_rps

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[128s])) / 128
```
