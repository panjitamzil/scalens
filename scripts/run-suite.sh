#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
scenario=${1:?cpu or database required}
root=${2:-results/$scenario}
mode=${3:-full}
case "$mode" in
  smoke) settings=scenarios/smoke.env ;;
  calibration|full) settings="scenarios/$scenario.env" ;;
  *) echo "Mode must be smoke, calibration or full" >&2; exit 1 ;;
esac
case "$scenario" in
  cpu) arms=(control scale hpa) ;;
  database) arms=(healthy control scale index hpa) ;;
  *) echo "Invalid scenario" >&2; exit 1 ;;
esac
if [[ "$mode" != full ]]; then
  if [[ "$scenario" == cpu ]]; then arms=(control scale); else arms=(control index); fi
fi
# Values already set in the environment win over the scenario file.
while IFS='=' read -r key value; do
  if [[ -n "$key" && -z "${!key:-}" ]]; then export "$key=$value"; fi
done < "$settings"
repeats=${REPEATS:-3}
if [[ "$mode" == calibration ]]; then repeats=${REPEATS:-1}; fi
[[ "$repeats" =~ ^[1-9][0-9]*$ && "$repeats" -le 10 ]] || { echo "REPEATS must be 1..10" >&2; exit 1; }
export EXPERIMENT_MODE="$mode"
failed=0
for ((block=1; block<=repeats; block++)); do
  for ((i=0; i<${#arms[@]}; i++)); do
    arm=${arms[$(((i+block-1)%${#arms[@]}))]}
    if ! bash scripts/run-experiment.sh "$scenario" "$arm" "$block" "$root"; then
      echo "run failed: $scenario $arm block $block" >&2
      failed=1
    fi
  done
done
exit "$failed"
