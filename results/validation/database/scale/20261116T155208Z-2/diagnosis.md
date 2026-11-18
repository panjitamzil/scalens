# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-16T15:53:08Z — 2026-11-16T15:55:16Z.

Status: **suspected** (suspected causes are not verified).

## Observations

- Error rate 0.989 exceeds configured 0.010
- SLO bad-event rate 1.000 exceeds 0.050

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 9.7268 | seconds |
| cpu_near_limit_fraction | 0 | fraction_of_samples |
| cpu_ratio | 0.077512 | ratio |
| cpu_throttled_ratio | 0.00041116 | ratio |
| db_server_mean | 2.2373 | seconds |
| error_rate | 0.98916 | ratio |
| failure_p50 | 7.5 | seconds |
| failure_p95 | 9.75 | seconds |
| failure_p99 | 9.95 | seconds |
| query_p95 | 4.4148 | seconds |
| requests | 2267.2 | requests |
| scrape_health | 1 | ratio |
| slo_bad_rate | 0.99955 | ratio |
| slo_bucket | 2 | ratio |
| success_p50 | 6.6667 | seconds |
| success_p95 | 9.6667 | seconds |
| success_p99 | 9.9333 | seconds |
| success_rps | 0.192 | requests/second |
| successes | 24.577 | requests |
| total_rps | 17.712 | requests/second |

## Suspect: slow_database_operation

Operations: [OrderRepository.CountRecentOrders OrderRepository.SaveOrder]

- Evidence: `query_p95`
- Evidence: `db_server_mean`
- Evidence: `trace:85d4049d4f704cf07f9413dd7afd078`
- Evidence: `trace:3b659b7d480ffc8104b277f78f6922fa`
- Evidence: `trace:99cb73258de2ea10cbf52aeb9cf5efb`
- Evidence: `trace:37a60197b6d2d47a01338061dfb8b247`
- Evidence: `trace:2447fbbfced466b4126b712e884686ff`
- Evidence: `trace:a9d0322afd9d89214e5189bfe0cd3f2`
- Evidence: `trace:98a88444bdabf3b0afa3e953b4ee924`
- Evidence: `trace:2fb77b17c55005ba027585b1c785f604`
- Evidence: `trace:17d331d16558be28066bdb48c057734c`
- Evidence: `trace:747467979510acf10f8c2840ebc8a580`
- Evidence: `trace:65ae082fdf69de4a08e1e55303efc5b8`
- Evidence: `trace:1928765bdea1985908b4ea207d484585`
- Evidence: `trace:6d1c6f7cb7ae6e4f1706ada9c5aec328`
- Evidence: `trace:561a773f161c1f231614790849dd5f34`
- Evidence: `trace:6bf939076b27ba5a1402ad1e2d9587e1`
- Evidence: `trace:65123d4a72b661940684b5942445cc59`
- Evidence: `trace:a9e46145d70f86870da224088e862448`
- Evidence: `trace:3de23651f52c2f2a154af0913f4c06c4`
- Evidence: `trace:4ff446f560e5fa080ec444b369b9d63b`
- Evidence: `trace:3a973cb50d77ee6044c7ff47fe16d26`
- Evidence: `trace:5163d711f202c0b106063233523195c4`
- Evidence: `trace:973f9c366837c53406d89718516ca1ad`
- Limitation: Statement time includes execution, lock and I/O waiting
- Limitation: Neither a missing index nor PostgreSQL saturation is proven by slow statements
- Limitation: Operations are ordered by sampled time; a saturated database slows unrelated statements too
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
