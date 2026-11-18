## cpu: control → scale

Status: **improved** (evidence: includes_flagged_runs).

| Measurement | Control | Candidate |
|---|---:|---:|
| Runs used | 3 | 3 |
| Median successful p95 (ms) | 8781.164 | 2587.496 |
| Median successful requests/s | 1.542 | 19.424 |
| Median error rate | 0.9120 | 0.0000 |
| Median SLO bad-event rate | 0.9964 | 0.1762 |

- Exploratory comparison; no statistical significance is claimed
- A diagnosis can be correct even when the intervention is ineffective
- Practical thresholds: 10% latency, 5% throughput, 1 percentage point error/SLO guardrail; 5 points when both arms fail most requests
- Runs flagged for host contention or dropped iterations are included; treat the outcome as indicative and repeat on a machine with headroom
- Where k6 dropped iterations the offered load was not reached and latency is understated

## cpu: control → hpa

Status: **improved** (evidence: includes_flagged_runs).

| Measurement | Control | Candidate |
|---|---:|---:|
| Runs used | 3 | 3 |
| Median successful p95 (ms) | 8781.164 | 7480.982 |
| Median successful requests/s | 1.542 | 15.285 |
| Median error rate | 0.9120 | 0.1848 |
| Median SLO bad-event rate | 0.9964 | 0.4740 |

- Exploratory comparison; no statistical significance is claimed
- A diagnosis can be correct even when the intervention is ineffective
- Practical thresholds: 10% latency, 5% throughput, 1 percentage point error/SLO guardrail; 5 points when both arms fail most requests
- Runs flagged for host contention or dropped iterations are included; treat the outcome as indicative and repeat on a machine with headroom
- Where k6 dropped iterations the offered load was not reached and latency is understated

## cpu-v2: control → scale

Status: **improved** (evidence: includes_flagged_runs).

| Measurement | Control | Candidate |
|---|---:|---:|
| Runs used | 3 | 3 |
| Median successful p95 (ms) | 7258.138 | 58.606 |
| Median successful requests/s | 10.540 | 20.002 |
| Median error rate | 0.4182 | 0.0000 |
| Median SLO bad-event rate | 0.9379 | 0.0008 |

- Exploratory comparison; no statistical significance is claimed
- A diagnosis can be correct even when the intervention is ineffective
- Practical thresholds: 10% latency, 5% throughput, 1 percentage point error/SLO guardrail; 5 points when both arms fail most requests
- Runs flagged for host contention or dropped iterations are included; treat the outcome as indicative and repeat on a machine with headroom
- Where k6 dropped iterations the offered load was not reached and latency is understated

## cpu-v2: control → hpa

Status: **improved** (evidence: includes_flagged_runs).

| Measurement | Control | Candidate |
|---|---:|---:|
| Runs used | 3 | 3 |
| Median successful p95 (ms) | 7258.138 | 1288.108 |
| Median successful requests/s | 10.540 | 19.666 |
| Median error rate | 0.4182 | 0.0000 |
| Median SLO bad-event rate | 0.9379 | 0.1073 |

- Exploratory comparison; no statistical significance is claimed
- A diagnosis can be correct even when the intervention is ineffective
- Practical thresholds: 10% latency, 5% throughput, 1 percentage point error/SLO guardrail; 5 points when both arms fail most requests
- Runs flagged for host contention or dropped iterations are included; treat the outcome as indicative and repeat on a machine with headroom
- Where k6 dropped iterations the offered load was not reached and latency is understated

## database: control → scale

Status: **no_material_change_observed** (evidence: includes_flagged_runs).

| Measurement | Control | Candidate |
|---|---:|---:|
| Runs used | 3 | 3 |
| Median successful p95 (ms) | 5751.063 | 6940.248 |
| Median successful requests/s | 0.094 | 0.023 |
| Median error rate | 0.9947 | 0.9987 |
| Median SLO bad-event rate | 0.9973 | 0.9996 |

- Exploratory comparison; no statistical significance is claimed
- A diagnosis can be correct even when the intervention is ineffective
- Practical thresholds: 10% latency, 5% throughput, 1 percentage point error/SLO guardrail; 5 points when both arms fail most requests
- Runs flagged for host contention or dropped iterations are included; treat the outcome as indicative and repeat on a machine with headroom
- Where k6 dropped iterations the offered load was not reached and latency is understated
- Both arms failed most requests; only error and SLO rates are compared

## database: control → hpa

Status: **no_material_change_observed** (evidence: includes_flagged_runs).

| Measurement | Control | Candidate |
|---|---:|---:|
| Runs used | 3 | 3 |
| Median successful p95 (ms) | 5751.063 | 5158.126 |
| Median successful requests/s | 0.094 | 0.055 |
| Median error rate | 0.9947 | 0.9969 |
| Median SLO bad-event rate | 0.9973 | 0.9978 |

- Exploratory comparison; no statistical significance is claimed
- A diagnosis can be correct even when the intervention is ineffective
- Practical thresholds: 10% latency, 5% throughput, 1 percentage point error/SLO guardrail; 5 points when both arms fail most requests
- Runs flagged for host contention or dropped iterations are included; treat the outcome as indicative and repeat on a machine with headroom
- Where k6 dropped iterations the offered load was not reached and latency is understated
- Both arms failed most requests; only error and SLO rates are compared

## database: control → index

Status: **improved** (evidence: includes_flagged_runs).

| Measurement | Control | Candidate |
|---|---:|---:|
| Runs used | 3 | 3 |
| Median successful p95 (ms) | 5751.063 | 27.712 |
| Median successful requests/s | 0.094 | 19.998 |
| Median error rate | 0.9947 | 0.0000 |
| Median SLO bad-event rate | 0.9973 | 0.0000 |

- Exploratory comparison; no statistical significance is claimed
- A diagnosis can be correct even when the intervention is ineffective
- Practical thresholds: 10% latency, 5% throughput, 1 percentage point error/SLO guardrail; 5 points when both arms fail most requests
- Runs flagged for host contention or dropped iterations are included; treat the outcome as indicative and repeat on a machine with headroom
- Where k6 dropped iterations the offered load was not reached and latency is understated

## database: healthy → index

Status: **no_material_change_observed** (evidence: includes_flagged_runs).

| Measurement | Control | Candidate |
|---|---:|---:|
| Runs used | 3 | 3 |
| Median successful p95 (ms) | 27.191 | 27.712 |
| Median successful requests/s | 20.004 | 19.998 |
| Median error rate | 0.0000 | 0.0000 |
| Median SLO bad-event rate | 0.0000 | 0.0000 |

- Exploratory comparison; no statistical significance is claimed
- A diagnosis can be correct even when the intervention is ineffective
- Practical thresholds: 10% latency, 5% throughput, 1 percentage point error/SLO guardrail; 5 points when both arms fail most requests
- Runs flagged for host contention or dropped iterations are included; treat the outcome as indicative and repeat on a machine with headroom
- Where k6 dropped iterations the offered load was not reached and latency is understated

