#!/usr/bin/env bash
set -euo pipefail
for tool in docker kind kubectl go k6 jq python3; do
  command -v "$tool" >/dev/null || { echo "Missing tool: $tool" >&2; exit 1; }
done
docker info --format '{"runtime":"{{.OperatingSystem}}","architecture":"{{.Architecture}}","cpus":{{.NCPU}},"memory_bytes":{{.MemTotal}}}'
docker context show
go version
kind version
k6 version
df -h .
python3 "$(dirname "$0")/host-health.py" snapshot
if kind get clusters 2>/dev/null | grep -qx scalens; then
  kubectl --context kind-scalens get nodes
  kubectl --context kind-scalens -n scalens get pods
  kubectl --context kind-scalens top nodes
fi
