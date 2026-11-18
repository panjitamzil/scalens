#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
scenario=${1:?usage: run-experiment.sh cpu|database arm block output-directory}
arm=${2:?arm required}
block=${3:-1}
root=${4:-results/experiment}
case "$scenario" in cpu|database) ;; *) echo "Invalid scenario" >&2; exit 1;; esac
case "$arm" in healthy|control|scale|index|hpa) ;; *) echo "Invalid arm" >&2; exit 1;; esac
rate=${RATE:-5}
seconds=${SECONDS_PER_RUN:-180}
warmup=${WARMUP_SECONDS:-60}
rows=${SEED_ROWS:-200000}
profile=${PROFILE:-constant}
work=${CPU_WORK:-250000}
replicas=${SCALE_REPLICAS:-2}
for value in "$rate" "$seconds" "$warmup" "$rows" "$work" "$block" "$replicas"; do
  [[ "$value" =~ ^(0|[1-9][0-9]*)$ ]] || { echo "Numeric settings must be nonnegative integers without leading zeros" >&2; exit 1; }
done
[[ "$seconds" -ge 30 && "$rate" -ge 1 ]] || { echo "Use at least 30 seconds and 1 request/s" >&2; exit 1; }
[[ "$seconds" -le 900 && "$rate" -le 100 && "$rows" -ge 1 && "$rows" -le 2000000 && "$work" -le 10000000 && "$warmup" -le 120 ]] || { echo "Settings exceed the bounded local lab profile" >&2; exit 1; }
case "$profile" in constant|step) ;; *) echo "PROFILE must be constant or step" >&2; exit 1;; esac
[[ "$replicas" -ge 2 && "$replicas" -le 4 ]] || { echo "SCALE_REPLICAS must be 2..4" >&2; exit 1; }
[[ -x bin/dbys ]] || { mkdir -p bin; go build -o bin/dbys ./cmd/dbys; }
k=(kubectl --context kind-scalens -n scalens)
id="$(date -u +%Y%m%dT%H%M%SZ)-${block}"
dir="$root/$arm/$id"
mkdir -p "$dir"
monitor=""
finish() {
  status=$?
  trap - EXIT
  if [[ -n "$monitor" ]]; then
    kill "$monitor" 2>/dev/null || true
    wait "$monitor" 2>/dev/null || true
  fi
  if [[ "$status" -ne 0 ]]; then
    jq -n --argjson code "$status" --arg scenario "$scenario" --arg arm "$arm" \
      '{status:"aborted",exit_code:$code,scenario:$scenario,arm:$arm}' >"$dir/failure.json"
  fi
  bash scripts/collect-diagnostics.sh "$dir/diagnostics" || true
  exit "$status"
}
trap finish EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
for attempt in 1 2 3; do
  if bash scripts/preflight.sh >"$dir/preflight.txt" 2>&1; then break; fi
  [[ "$attempt" -lt 3 ]] || { echo "preflight failed; see $dir/preflight.txt" >&2; exit 1; }
  sleep 15
done
"${k[@]}" delete hpa order-service --ignore-not-found
"${k[@]}" scale deployment/order-service --replicas=1
if [[ "$scenario" != cpu ]]; then work=0; fi
"${k[@]}" set env deployment/order-service "ORDER_SYNTHETIC_WORK=$work"
if [[ "$arm" == scale ]]; then "${k[@]}" scale deployment/order-service --replicas="$replicas"; fi
"${k[@]}" rollout status deployment/order-service --timeout=180s
"${k[@]}" exec -i deployment/postgres -- psql -U scalens -d scalens -v ON_ERROR_STOP=1 < database/migrations/001_orders.sql >"$dir/seed.log" 2>&1
"${k[@]}" exec -i deployment/postgres -- psql -U scalens -d scalens -v ON_ERROR_STOP=1 -v "rows=$rows" < database/seed.sql >>"$dir/seed.log"
sql='CREATE INDEX IF NOT EXISTS orders_customer_created_idx ON orders(customer_id,created_at);'
if [[ "$scenario" == database && "$arm" != index && "$arm" != healthy ]]; then
  sql='DROP INDEX IF EXISTS orders_customer_created_idx;'
fi
"${k[@]}" exec deployment/postgres -- psql -U scalens -d scalens -v ON_ERROR_STOP=1 -c "$sql ANALYZE orders;" >"$dir/index.log"
"${k[@]}" exec deployment/postgres -- psql -U scalens -d scalens -Atc "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) SELECT count(*) FROM orders WHERE customer_id=42 AND created_at >= TIMESTAMPTZ '2026-01-01' AND created_at < TIMESTAMPTZ '2026-02-01';" >"$dir/explain.json"
if [[ "$warmup" -gt 0 ]]; then
  RATE="$rate" WORKLOAD_SECONDS="$warmup" SUMMARY_PATH="$dir/warmup.json" k6 run --quiet workloads/k6/orders.js >"$dir/warmup.log" 2>&1
fi
if [[ "$arm" == hpa ]]; then
  kubectl --context kind-scalens apply -f deploy/kubernetes/hpa.yaml >"$dir/hpa-apply.txt"
  cp deploy/kubernetes/hpa.yaml "$dir/hpa.yaml"
fi
"${k[@]}" get deployments,hpa,pods -o json >"$dir/kubernetes-before.json"
"${k[@]}" exec -i deployment/postgres -- psql -U scalens -d scalens -At < database/diagnostics/snapshot.sql >"$dir/postgres-before.json"
python3 scripts/host-health.py snapshot >"$dir/host-samples.jsonl"
(
  while true; do
    date -u +%Y-%m-%dT%H:%M:%SZ
    sleep 10
    python3 scripts/host-health.py snapshot >>"$dir/host-samples.jsonl" || true
    kubectl --context kind-scalens --request-timeout=5s top nodes || true
    "${k[@]}" --request-timeout=5s get pods,hpa -o wide || true
  done
) >"$dir/runtime.log" 2>&1 &
monitor=$!
start=$(date -u +%Y-%m-%dT%H:%M:%SZ)
set +e
RATE="$rate" WORKLOAD_SECONDS="$seconds" PROFILE="$profile" SUMMARY_PATH="$dir/k6-summary.json" k6 run --quiet --out "json=$dir/k6-series.jsonl" workloads/k6/orders.js >"$dir/k6.log" 2>&1
code=$?
set -e
end=$(date -u +%Y-%m-%dT%H:%M:%SZ)
kill "$monitor" 2>/dev/null || true
wait "$monitor" 2>/dev/null || true
monitor=""
python3 scripts/host-health.py snapshot >>"$dir/host-samples.jsonl"
python3 scripts/host-health.py assess "$dir/host-samples.jsonl" >"$dir/host-health.json"
"${k[@]}" exec -i deployment/postgres -- psql -U scalens -d scalens -At < database/diagnostics/snapshot.sql >"$dir/postgres-after.json"
"${k[@]}" get deployments,hpa,pods -o json >"$dir/kubernetes-after.json"
"${k[@]}" get events -o json >"$dir/events.json"
validity=valid
reasons='[]'
if [[ "$code" -ne 0 || ! -f "$dir/k6-summary.json" ]]; then
  echo "k6 failed; see $dir/k6.log" >&2
  exit 1
fi
if jq -e '.dropped_iterations > 0' "$dir/k6-summary.json" >/dev/null; then
  validity=load_not_achieved; reasons='["k6 dropped iterations"]'
fi
host_status=$(jq -r .status "$dir/host-health.json")
if [[ "$host_status" != valid ]]; then
  validity="$host_status"
  reasons=$(jq -c --argjson previous "$reasons" '$previous + .reasons' "$dir/host-health.json")
fi
if jq -e --slurpfile before "$dir/kubernetes-before.json" '
  ($before[0].items | map(select(.kind=="Pod")) | map({key:.metadata.uid,value:([.status.containerStatuses[]?.restartCount]|add//0)}) | from_entries) as $restarts |
  [.items[] | select(.kind=="Pod" and .metadata.deletionTimestamp==null) |
    select(([.status.containerStatuses[]? | select(.ready != true)]|length)>0 or
      ([.status.containerStatuses[]?.restartCount]|add//0)>($restarts[.metadata.uid]//0))] | length > 0
  ' "$dir/kubernetes-after.json" >/dev/null; then
  validity=contended; reasons=$(jq -c '.+["pod not ready or restarted; inspect artifacts"]' <<<"$reasons")
fi
git rev-parse HEAD >"$dir/revision.txt" 2>/dev/null || echo "uncommitted" >"$dir/revision.txt"
git diff --no-ext-diff >"$dir/working.diff"
git status --porcelain >"$dir/working-status.txt"
find cmd internal services database deploy workloads scripts -type f -not -name '.DS_Store' -exec shasum -a 256 {} \; | sort >"$dir/source-sha256.txt"
docker image ls --digests --format '{{.Repository}}:{{.Tag}} {{.Digest}}' >"$dir/images.txt"
docker info --format '{"architecture":"{{.Architecture}}","cpus":{{.NCPU}},"memory_bytes":{{.MemTotal}}}' >"$dir/runtime-capacity.json"
resource_profile=$(jq -Sc --arg host "$(uname -s)-$(uname -m)" --arg work "$work" \
  --slurpfile capacity "$dir/runtime-capacity.json" \
  '{host:$host,work:$work,capacity:$capacity[0],deployments:[.items[]|select(.kind=="Deployment")|
    {name:.metadata.name,containers:[.spec.template.spec.containers[]|{name,image,resources}]}]|sort_by(.name)}' \
  "$dir/kubernetes-before.json" | shasum -a 256 | awk '{print $1}')
sleep "${INGEST_WAIT:-15}"
curl -fsS localhost:9090/api/v1/query --data-urlencode "time=$end" \
  --data-urlencode "query=sum by(pod) (increase(http_server_requests_total{namespace=\"scalens\",service=\"order-service\",route=\"/orders\"}[${seconds}s]))" \
  | jq '[.data.result[] | {pod:.metric.pod, requests:(.value[1]|tonumber)}]' >"$dir/pod-requests.json"
bin/dbys diagnose --start "$start" --end "$end" --offered-rate "$rate" --workload "orders-seed-v1" \
  --snapshot-out "$dir/telemetry.json" --format json --out "$dir/diagnosis.json"
bin/dbys diagnose --input "$dir/telemetry.json" --out "$dir/diagnosis.md"
if jq -e '[.missing_evidence[] | select(test("^traces?[ :]|^no traces"))] | length > 0' "$dir/diagnosis.json" >/dev/null; then
  if [[ "$validity" == valid ]]; then validity=telemetry_incomplete; fi
  reasons=$(jq -c '.+["trace retrieval incomplete"]' <<<"$reasons")
fi
jq -n --arg schema_version 1 --arg id "$id" --arg scenario "$scenario" --arg arm "$arm" --argjson block "$block" \
  --argjson rate "$rate" --argjson seconds "$seconds" --argjson rows "$rows" --arg profile "$profile" \
  --arg resource_profile "$resource_profile" --arg mode "${EXPERIMENT_MODE:-custom}" \
  --argjson cpu_work "$work" --argjson warmup "$warmup" \
  --argjson preallocated_vus "${PREALLOCATED_VUS:-20}" --argjson max_vus "${MAX_VUS:-$((rate * 15 > 20 ? rate * 15 : 20))}" \
  --argjson scale_replicas "$replicas" --arg validity "$validity" --argjson reasons "$reasons" \
  --arg start "$start" --arg end "$end" --slurpfile stats "$dir/k6-summary.json" \
  '{schema_version:$schema_version,id:$id,scenario:$scenario,arm:$arm,block:$block,rate:$rate,seconds:$seconds,rows:$rows,profile:$profile,resource_profile:$resource_profile,mode:$mode,cpu_work:$cpu_work,warmup_seconds:$warmup,preallocated_vus:$preallocated_vus,max_vus:$max_vus,scale_replicas:$scale_replicas,validity:$validity,reasons:$reasons,start:$start,end:$end,stats:($stats[0]|del(.raw))}' >"$dir/run.json"
echo "$dir ($validity)"
