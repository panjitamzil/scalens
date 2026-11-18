# Diagnosis: order-service

Target: `order-service /orders` in `scalens`.

Window: 2026-11-16T07:14:54Z — 2026-11-16T07:47:16Z.

Status: **suspected** (suspected causes are not verified).

## Observations

- Error rate 0.991 exceeds configured 0.010
- SLO bad-event rate 1.000 exceeds 0.050

## Measurements

| Signal | Value | Unit |
|---|---:|---|
| acquire_p95 | 7.7083 | seconds |
| cpu_near_limit_fraction | 0.66667 | fraction_of_samples |
| cpu_ratio | 0.87182 | ratio |
| cpu_throttled_ratio | 0.89353 | ratio |
| db_server_mean | 0.015796 | seconds |
| error_rate | 0.9905 | ratio |
| failure_p50 | 7.0866 | seconds |
| failure_p95 | 10 | seconds |
| failure_p99 | 10 | seconds |
| query_p95 | 5 | seconds |
| requests | 2007.3 | requests |
| scrape_health | 0.56 | ratio |
| slo_bad_rate | 1 | ratio |
| slo_bucket | 1 | ratio |
| success_p50 | 4.1964 | seconds |
| success_p95 | 9.05 | seconds |
| success_p99 | 9.81 | seconds |
| success_rps | 0.0098194 | requests/second |
| successes | 19.069 | requests |
| total_rps | 1.0336 | requests/second |

## Suspect: cpu_saturation

Operations: []

- Evidence: `cpu_ratio`
- Evidence: `cpu_near_limit_fraction`
- Evidence: `cpu_throttled_ratio`
- Limitation: Container CPU evidence does not identify a CPU-expensive function
- Limitation: Host contention must be checked independently
- Experiment: Compare one and two replicas at the same offered load with host headroom
- Experiment: Inspect CPU profiles if local computation remains unclear

## Suspect: slow_database_operation

Operations: [OrderRepository.CountRecentOrders]

- Evidence: `query_p95`
- Evidence: `db_server_mean`
- Evidence: `trace:13d7c18a3bd8227d0e464fc8636a5112`
- Limitation: Statement time includes execution, lock and I/O waiting
- Limitation: Neither a missing index nor PostgreSQL saturation is proven by slow statements
- Limitation: Connection acquisition is slower than query execution; pool waiting may also contribute
- Experiment: Inspect EXPLAIN (ANALYZE, BUFFERS) and pg_stat_statements for the operation
- Experiment: Compare the relevant index change with application scale-out

Traces analyzed: 10; incomplete: 0; candidates: 10.

Sampling: Tempo bounded first-N search, split across time windows; not random or representative.

## Missing evidence and warnings

- application scrape coverage below 95% or unknown
- Scrapes were missed during the window; counters are under-sampled and an overloaded target can cause this
- Multiple supported suspects; evaluate interventions separately
- Sampled traces are selected evidence, not a population latency estimate

## Metric provenance

### acquire_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_acquire_duration_seconds_bucket{service="order-service",namespace="scalens",outcome="success"}[1942s])))
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
sum(increase(container_cpu_cfs_throttled_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[1942s])) / sum(increase(container_cpu_cfs_periods_total{namespace="scalens",container="order-service",pod=~"order-service-.*"}[1942s]))
```

### db_server_mean

```promql
max((sum by(queryid) (increase(pg_stat_statements_seconds_total[1942s])) / (sum by(queryid) (increase(pg_stat_statements_calls_total[1942s])) > 0)) and topk(1, sum by(queryid) (increase(pg_stat_statements_seconds_total[1942s]))))
```

### error_rate

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[1942s])) - (sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[1942s])) or vector(0))) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[1942s]))
```

### failure_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[1942s])))
```

### failure_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[1942s])))
```

### failure_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="failure"}[1942s])))
```

### query_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(db_client_query_duration_seconds_bucket{service="order-service",namespace="scalens",operation="OrderRepository.CountRecentOrders",outcome="success"}[1942s])))
```

### requests

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[1942s]))
```

### scrape_health

```promql
min(avg_over_time(up{job="application",namespace="scalens"}[1942s]))
```

### slo_bad_rate

```promql
1 - ((sum(increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"}[1942s])) or vector(0)) / sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[1942s])))
```

### slo_bucket

```promql
count(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success",le="0.5"})
```

### success_p50

```promql
histogram_quantile(0.5, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[1942s])))
```

### success_p95

```promql
histogram_quantile(0.95, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[1942s])))
```

### success_p99

```promql
histogram_quantile(0.99, sum by(le) (increase(http_server_request_duration_seconds_bucket{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[1942s])))
```

### success_rps

```promql
(sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[1942s])) or vector(0)) / 1942
```

### successes

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens",outcome="success"}[1942s]))
```

### total_rps

```promql
sum(increase(http_server_requests_total{service="order-service",route="/orders",method="POST",namespace="scalens"}[1942s])) / 1942
```
