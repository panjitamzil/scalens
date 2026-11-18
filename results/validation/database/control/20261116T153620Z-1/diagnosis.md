# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-16T15:37:07Z — 2026-11-16T15:39:16Z.

Status: **suspected** (suspected causes are not verified).

## Observations

- Error rate 0.998 exceeds configured 0.010
- SLO bad-event rate 0.999 exceeds 0.050

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 9.7413 | seconds |
| cpu_near_limit_fraction | 0 | fraction_of_samples |
| cpu_ratio | 0.13739 | ratio |
| cpu_throttled_ratio | 0.002669 | ratio |
| db_server_mean | 1.3531 | seconds |
| error_rate | 0.99772 | ratio |
| failure_p50 | 7.5 | seconds |
| failure_p95 | 9.75 | seconds |
| failure_p99 | 9.95 | seconds |
| query_p95 | 2.3832 | seconds |
| requests | 2267.2 | requests |
| scrape_health | 1 | ratio |
| slo_bad_rate | 0.99909 | ratio |
| slo_bucket | 1 | ratio |
| success_p50 | 0.625 | seconds |
| success_p95 | 8.75 | seconds |
| success_p99 | 9.75 | seconds |
| success_rps | 0.039999 | requests/second |
| successes | 5.1598 | requests |
| total_rps | 17.575 | requests/second |

## Suspect: slow_database_operation

Operations: [OrderRepository.CountRecentOrders OrderRepository.SaveOrder]

- Evidence: `query_p95`
- Evidence: `db_server_mean`
- Evidence: `trace:2518776a4f37ee481576b390ac4eaf14`
- Evidence: `trace:422089d644f300a312e52402bb038025`
- Evidence: `trace:f743d2424be9f983080ac2cbc6f8ed10`
- Evidence: `trace:67a8c5ec0ffa5c050ab5e262c1d1e926`
- Evidence: `trace:e555f7e96f2748c5184f982ee57d522d`
- Evidence: `trace:c984053b2b5dbd0f13155fd0a822e0c2`
- Evidence: `trace:6f7f53a041f1549b18b974034c993033`
- Evidence: `trace:b44744ee9058daaf02868cf97e23db10`
- Evidence: `trace:27536b318aed83611771909cb4943630`
- Evidence: `trace:424fbd22361655f00fede3c2baeab47c`
- Evidence: `trace:a70938a3b8625da90eaef17849614b2a`
- Evidence: `trace:cc670fa94b6e303b0f6c882e6c11ddc5`
- Evidence: `trace:747d54ef8b764070c97b3ec7781b5d5`
- Evidence: `trace:7027cafcebab62e07c29814703052d2`
- Evidence: `trace:c793b6ed2072b71117ed5017411f908`
- Evidence: `trace:5f69a0bd3ee51d115c1f3bd6101e2c7`
- Evidence: `trace:129c89aaf8c3d2e113450e867d89ead2`
- Evidence: `trace:120c46cac949b1cc10b2d6c94befd95e`
- Evidence: `trace:4417e4744eecce180ef9cf75c94496bd`
- Evidence: `trace:40db61860475c01510770732c2ce9da1`
- Evidence: `trace:43b7d54f7c54590f055a94db310d5f55`
- Evidence: `trace:78c9f409a56936771329bd0aef774366`
- Evidence: `trace:372f5c27347c481106795d8e8d3afc32`
- Limitation: Statement time includes execution, lock and I/O waiting
- Limitation: Neither a missing index nor PostgreSQL saturation is proven by slow statements
- Limitation: Operations are ordered by sampled time; a saturated database slows unrelated statements too
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
