# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-16T16:11:13Z — 2026-11-16T16:13:21Z.

Status: **suspected** (suspected causes are not verified).

## Observations

- Error rate 0.997 exceeds configured 0.010
- SLO bad-event rate 0.998 exceeds 0.050

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 9.7396 | seconds |
| cpu_near_limit_fraction | 0 | fraction_of_samples |
| cpu_ratio | 0.15004 | ratio |
| cpu_throttled_ratio | 0.003663 | ratio |
| db_server_mean | 1.4495 | seconds |
| error_rate | 0.9968 | ratio |
| failure_p50 | 7.5 | seconds |
| failure_p95 | 9.75 | seconds |
| failure_p99 | 9.95 | seconds |
| query_p95 | 2.4342 | seconds |
| requests | 2240.5 | requests |
| scrape_health | 1 | ratio |
| slo_bad_rate | 0.99771 | ratio |
| slo_bucket | 1 | ratio |
| success_p50 | 0.3125 | seconds |
| success_p95 | 9.125 | seconds |
| success_p99 | 9.825 | seconds |
| success_rps | 0.056 | requests/second |
| successes | 7.168 | requests |
| total_rps | 17.504 | requests/second |

## Suspect: slow_database_operation

Operations: [OrderRepository.CountRecentOrders]

- Evidence: `query_p95`
- Evidence: `db_server_mean`
- Evidence: `trace:c02f6f6e883c2040c37a9eb50a89f08`
- Evidence: `trace:479a473caf6d686f0ea1b0ec1304a56d`
- Evidence: `trace:45a2eec29daaca17191e74ae30d1af5f`
- Evidence: `trace:e1ff90dfb6f40f41503511e72d88b11`
- Evidence: `trace:2408296a26fa45750a85a4077d3cf2de`
- Evidence: `trace:f2e970b7cac788e0bacb6df9b917a2d`
- Evidence: `trace:3484fc78e2e42ef415bdab734ba9efd1`
- Evidence: `trace:280d7203a1330e38170424774acd44c2`
- Evidence: `trace:4a4202b9bf0b74520d435ed98070a923`
- Evidence: `trace:328748225a67146a00c757cffebc1ae6`
- Evidence: `trace:1d390141775d864500ab99165f751d38`
- Evidence: `trace:21a1b8d6ec756035082a53509734975b`
- Evidence: `trace:1cbe7974b18d856606085a0e3b67fa2e`
- Evidence: `trace:1ac391c5f42cf05809000322dd7811f9`
- Evidence: `trace:19486ef68d2441d507196144924ba37e`
- Evidence: `trace:1a62cd1844ce48110f053f117cf0bbcc`
- Evidence: `trace:1436bb5325bf933b0dac84415814cdab`
- Evidence: `trace:42e0f3e251c726c50e474f8d16db9a9b`
- Evidence: `trace:1613f3a71dea3e1d1531a92d8356b09d`
- Evidence: `trace:1b408321f0beb45317e87520dc0df502`
- Evidence: `trace:3fa04a14bb7c319e05c354f749def46a`
- Evidence: `trace:3cc2114e0bb90dba09bfbf4fbd1174e8`
- Evidence: `trace:44c3665b69fb838617f9bbe14511a255`
- Evidence: `trace:4c9c1850b290e5c16d5260474b23444`
- Evidence: `trace:39bd3b748c1604c910da68afa4ce726d`
- Limitation: Statement time includes execution, lock and I/O waiting
- Limitation: Neither a missing index nor PostgreSQL saturation is proven by slow statements
- Limitation: Connection acquisition is slower than query execution; pool waiting may also contribute
- Experiment: Inspect EXPLAIN (ANALYZE, BUFFERS) and pg_stat_statements for the operation
- Experiment: Compare the relevant index change with application scale-out

Traces analyzed: 28; incomplete: 0; candidates: 30.

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
