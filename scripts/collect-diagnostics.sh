#!/usr/bin/env bash
set -uo pipefail
dir=${1:?diagnostic output directory required}
mkdir -p "$dir/logs"
k=(kubectl --context kind-scalens --request-timeout=5s)
"${k[@]}" get nodes -o wide >"$dir/nodes.txt" 2>&1 || true
"${k[@]}" -n scalens get pods,deploy,hpa -o json >"$dir/objects.json" 2>"$dir/objects-error.txt" || true
"${k[@]}" get events -A -o json >"$dir/events.json" 2>&1 || true
"${k[@]}" -n scalens describe pods >"$dir/pods.txt" 2>&1 || true
"${k[@]}" -n kube-system describe pods >"$dir/system-pods.txt" 2>&1 || true
for namespace in scalens kube-system; do
  while IFS= read -r pod; do
    [[ -n "$pod" ]] || continue
    "${k[@]}" -n "$namespace" logs "$pod" --all-containers --tail=200 --timestamps >"$dir/logs/$namespace-$pod.log" 2>&1 || true
    "${k[@]}" -n "$namespace" logs "$pod" --all-containers --previous --tail=100 --timestamps >"$dir/logs/$namespace-$pod-previous.log" 2>&1 || true
  done < <("${k[@]}" -n "$namespace" get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' 2>/dev/null)
done
curl -fsS --max-time 10 localhost:9090/api/v1/targets >"$dir/targets.json" 2>"$dir/targets-error.txt" || true
exit 0
