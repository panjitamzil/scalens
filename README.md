# ScaleNS — Diagnose Before You Scale

A slow service is not always short of replicas. ScaleNS is a small, reproducible lab and a CLI that follow one workflow:

**Observe → Diagnose → Recommend → Experiment → Verify**

It contains two Go services with PostgreSQL, a local Kubernetes environment with Prometheus and Tempo, two injectable bottlenecks, k6 workloads, and `dbys`, a rule-based CLI that reads metrics and traces and reports a *suspected* bottleneck with its evidence.

This is an engineering proof of concept. It is not an autoscaler, not a new algorithm, and not production ready. Measured results and their limits are in [docs/results.md](docs/results.md).

## What is in the repository

| Path | Content |
|---|---|
| `services/`, `cmd/order-service`, `cmd/payment-service` | Order and Payment services |
| `cmd/dbys`, `internal/` | Diagnosis CLI, telemetry clients, rules, comparison |
| `database/` | Migration, deterministic seed, statistics snapshot |
| `deploy/` | Dockerfile, Compose, kind, Kubernetes and observability manifests |
| `workloads/k6/` | Constant and step arrival-rate workloads |
| `scenarios/`, `scripts/` | Experiment settings, preflight, runner, evaluation |
| `docs/` | Research notes, architecture, telemetry contract, method, results, demo |

## Requirements

Docker (OrbStack or Docker Desktop), Go 1.27, kind, kubectl, k6, jq and Python 3 (standard library only). No cloud account is needed. The lab was built on an Apple M2 with 8 GB RAM; about 4 GB for the Docker VM is the practical minimum.

## Quick start

```sh
make test                 # unit tests
make compose-up           # Order on localhost:18080, PostgreSQL on 15432
make integration          # tests against the Compose database
make compose-down

make cluster-up           # kind cluster, images, Prometheus, Tempo, Metrics Server
make build
curl -s localhost:8080/orders -H 'Content-Type: application/json' \
  -d '{"customer_id":42,"items":[{"price":1000,"quantity":2}]}'
bin/dbys diagnose --last 2m
```

`make cluster-up` maps three ports to the host: Order on 8080, Prometheus on 9090 and the Tempo query API on 3200. Compose uses ports 18080, 18081 and 15432 so it can run next to the cluster and a local PostgreSQL.

## Diagnosing

```sh
# Capture a healthy reference window. The command refuses a window that is not healthy.
bin/dbys baseline capture --start 2026-11-20T10:00:00Z --end 2026-11-20T10:05:00Z --out healthy.json

# Diagnose an explicit window against it.
bin/dbys diagnose --start 2026-11-20T10:10:00Z --end 2026-11-20T10:15:00Z --baseline-file healthy.json

# Or use the last few minutes with only the latency SLO as reference.
bin/dbys diagnose --last 5m --slo 0.5 --format json
```

Use either `--start/--end` (RFC 3339) or `--last`. A baseline comes from `--baseline-file` or `--baseline-start/--baseline-end`, and may not overlap the diagnosed window. Without a baseline and with `--slo 0` the answer is `insufficient_evidence`.

| Status | Meaning |
|---|---|
| `suspected` | A symptom exists and at least one rule has supporting evidence |
| `no_anomaly` | A reference exists, data is complete, and nothing degraded |
| `insufficient_evidence` | Data, a reference, or cause evidence is missing |
| `conflicting_evidence` | Counters disagree with each other |

A report lists the PromQL used, values, trace IDs, missing data and the experiment that would test each suspect. A suspect is a hypothesis; only a before/after comparison shows whether a remediation works. Rules and their failure modes are in [docs/diagnosis.md](docs/diagnosis.md), required signals in [docs/telemetry-contract.md](docs/telemetry-contract.md). Backend addresses, thresholds and PromQL can be overridden with `--config` (see `dbys.example.json`).

## Experiments

Start small on your laptop:

```sh
make cluster-up           # once, if the cluster is not running
make smoke
```

The smoke profile uses 1,000 rows, 2 requests/s, 1,000 hash iterations, two preallocated k6 VUs, and one 45-second run per arm. It runs CPU control/scale and database control/index, without HPA. Expect about 5–7 minutes including warm-up, deployment changes and telemetry collection. It checks the workflow; it is not a bottleneck benchmark. `no_anomaly` is a normal result at this size. The observability stack is the same, so host contention may still be reported.

For only one scenario, or an explicit size override:

```sh
bash scripts/run-suite.sh database results/smoke-db smoke
RATE=3 SEED_ROWS=2000 bash scripts/run-suite.sh database results/smoke-db-larger smoke
```

For measurements after calibration:

```sh
make preflight            # tools, VM capacity, host memory pressure
make cpu                  # CPU saturation: 1 replica, 2 replicas, HPA
make database             # slow query: healthy, no index, no index + replicas, index, HPA
bash scripts/evaluate.sh results           # diagnosis versus injected fault
bin/dbys compare --control results/database/control --candidate results/database/index
```

Each run stores its settings, k6 summary, Kubernetes state, PostgreSQL plan and statistics, host samples, the telemetry snapshot and the diagnosis under `results/<scenario>/<arm>/<run>/`. Runs get a validity label; `compare` uses only `valid` runs unless `--include-flagged` is given, and then labels the outcome. The method is described in [docs/experiments.md](docs/experiments.md).

Numbers from this lab describe one laptop. They are not a benchmark of Kubernetes, Go or PostgreSQL.

The manual **Local lab experiments** GitHub Actions workflow supports `smoke`, `calibration` and `full` modes, with size overrides and downloadable diagnostics even on failure. Start with smoke on the runner, then calibrate its workload before collecting repeated measurements. See [docs/github-actions.md](docs/github-actions.md) for the exact sequence and offline debugging.

## Clean up

```sh
kind delete cluster --name scalens
```

## Scope

Included: two bottleneck classes (CPU saturation of a service, slow database statement), static replicas and CPU-based HPA, explainable reports.

Planned extensions, not implemented: retry amplification, backoff and circuit breakers, profiling, an OpenTelemetry Collector, custom-metric autoscaling. Grafana is optional and not installed.

Background reading and how it is used: [docs/research.md](docs/research.md). A walkthrough for presenting the lab: [docs/demo.md](docs/demo.md).

## License

MIT
