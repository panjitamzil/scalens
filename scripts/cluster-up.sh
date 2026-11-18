#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
for tool in docker kind kubectl; do command -v "$tool" >/dev/null; done
if ! kind get clusters 2>/dev/null | grep -qx scalens; then
  kind create cluster --name scalens --config deploy/kind/cluster.yaml --image kindest/node:v1.35.0 --wait 180s
fi
k=(kubectl --context kind-scalens)
docker build -f deploy/docker/Dockerfile --build-arg SERVICE=order-service -t scalens/order:local .
docker build -f deploy/docker/Dockerfile --build-arg SERVICE=payment-service -t scalens/payment:local .
kind load docker-image --name scalens scalens/order:local scalens/payment:local
"${k[@]}" apply -f deploy/kubernetes/app.yaml
"${k[@]}" -n scalens rollout status deployment/postgres --timeout=180s
"${k[@]}" -n scalens exec -i deployment/postgres -- psql -U scalens -d scalens -v ON_ERROR_STOP=1 < database/migrations/001_orders.sql
"${k[@]}" -n scalens create configmap prometheus-config --from-file=deploy/observability/prometheus.yaml --dry-run=client -o yaml | "${k[@]}" apply -f -
"${k[@]}" -n scalens create configmap tempo-config --from-file=deploy/observability/tempo.yaml --dry-run=client -o yaml | "${k[@]}" apply -f -
"${k[@]}" apply -f deploy/observability/stack.yaml
"${k[@]}" apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.8.0/components.yaml
"${k[@]}" -n kube-system patch deployment metrics-server --type=json -p='[{"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}]'
"${k[@]}" -n scalens rollout status deployment --timeout=240s
"${k[@]}" -n kube-system rollout status deployment/metrics-server --timeout=180s
curl --fail --retry 10 --retry-delay 2 http://localhost:8080/readyz
curl --fail http://localhost:9090/-/ready
curl --fail --retry 10 --retry-delay 3 http://localhost:3200/ready
