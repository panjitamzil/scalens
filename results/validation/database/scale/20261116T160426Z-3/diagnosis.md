# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-16T16:05:13Z — 2026-11-16T16:07:22Z.

Status: **suspected** (suspected causes are not verified).

## Observations

- Error rate 0.999 exceeds configured 0.010
- SLO bad-event rate 1.000 exceeds 0.050

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 9.7377 | seconds |
| cpu_near_limit_fraction | 0 | fraction_of_samples |
| cpu_ratio | 0.082745 | ratio |
| cpu_throttled_ratio | 0.00225 | ratio |
| db_server_mean | 2.6072 | seconds |
| error_rate | 0.99907 | ratio |
| failure_p50 | 7.5 | seconds |
| failure_p95 | 9.75 | seconds |
| failure_p99 | 9.95 | seconds |
| query_p95 | 4.8218 | seconds |
| requests | 2258.1 | requests |
| scrape_health | 1 | ratio |
| slo_bad_rate | 1 | ratio |
| slo_bucket | 2 | ratio |
| success_p50 | 5.1 | seconds |
| success_p95 | 9.51 | seconds |
| success_p99 | 9.902 | seconds |
| success_rps | 0.016333 | requests/second |
| successes | 2.107 | requests |
| total_rps | 17.505 | requests/second |

## Suspect: slow_database_operation

Operations: [OrderRepository.CountRecentOrders]

- Evidence: `query_p95`
- Evidence: `db_server_mean`
- Evidence: `trace:14db4236e8c422150d99976c92594cc9`
- Evidence: `trace:1a6eded3f3b786d02a753f42349401c`
- Evidence: `trace:5c539a31f710db3144c4e21615470b5`
- Evidence: `trace:11896edb0c05d630163567989cb81a54`
- Evidence: `trace:1f7516604bd2ecaa089ea879915586e3`
- Evidence: `trace:7af7e7d619dbf4f138492011eddbda2`
- Evidence: `trace:e766f953f72b9070b20df47d7ac18cb`
- Evidence: `trace:bcaea9d7c87e9cd074c94583c74002e`
- Evidence: `trace:170b2c399c9b5d0d01b27f0250dffc7b`
- Evidence: `trace:4e353e072456670183ad5360148da7`
- Evidence: `trace:10c30780fa8d68310fd726f834f35b8b`
- Evidence: `trace:7e4c9bd302360181623391a5697570a`
- Evidence: `trace:2320099bd94bcf7d1918f10a3ca1dd85`
- Evidence: `trace:8b9565b676614460d9a0aadc8b0bf51`
- Evidence: `trace:3ace0535f42f6e5b11445f3f4a534da5`
- Evidence: `trace:9ce0282ca0a296f176b5859c49121fc`
- Evidence: `trace:ccd09958aeffb8e018775e3d59ca37f2`
- Limitation: Statement time includes execution, lock and I/O waiting
- Limitation: Neither a missing index nor PostgreSQL saturation is proven by slow statements
- Limitation: Connection acquisition is slower than query execution; pool waiting may also contribute
- Experiment: Inspect EXPLAIN (ANALYZE, BUFFERS) and pg_stat_statements for the operation
- Experiment: Compare the relevant index change with application scale-out

Traces analyzed: 26; incomplete: 0; candidates: 30.

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
