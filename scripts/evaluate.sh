#!/usr/bin/env bash
set -euo pipefail
root=${1:?usage: evaluate.sh results-directory}
# Ground truth lives here and is never passed to the diagnosis engine.
expected() {
  case "$1/$2" in
    cpu/control) echo cpu_saturation ;;
    database/control|database/scale|database/hpa) echo slow_database_operation ;;
    database/healthy|database/index) echo none ;;
    *) echo unscored ;;
  esac
}
echo "| Scenario | Arm | Run | Validity | Injected fault | Status | Suspects | Verdict |"
echo "|---|---|---|---|---|---|---|---|"
correct=0; partial=0; wrong=0; abstain=0; unscored=0
while IFS= read -r run; do
  dir=$(dirname "$run")
  [[ -f "$dir/diagnosis.json" ]] || continue
  scenario=$(jq -r .scenario "$run"); arm=$(jq -r .arm "$run"); validity=$(jq -r .validity "$run")
  truth=$(expected "$scenario" "$arm")
  status=$(jq -r .status "$dir/diagnosis.json")
  suspects=$(jq -r '[.suspects[].category] | join(", ")' "$dir/diagnosis.json")
  count=$(jq '.suspects | length' "$dir/diagnosis.json")
  if [[ "$truth" == unscored ]]; then
    verdict=unscored; unscored=$((unscored+1))
  elif [[ "$truth" == none ]]; then
    if [[ "$count" -eq 0 && "$status" == no_anomaly ]]; then verdict=correct; correct=$((correct+1))
    elif [[ "$count" -eq 0 ]]; then verdict=abstain; abstain=$((abstain+1))
    else verdict="wrong (false positive)"; wrong=$((wrong+1)); fi
  elif [[ "$count" -eq 0 ]]; then
    if [[ "$status" == no_anomaly ]]; then verdict="wrong (missed)"; wrong=$((wrong+1))
    else verdict=abstain; abstain=$((abstain+1)); fi
  elif [[ "$suspects" == "$truth" ]]; then verdict=correct; correct=$((correct+1))
  elif [[ ", $suspects, " == *", $truth, "* ]]; then verdict="partial (extra suspect)"; partial=$((partial+1))
  else verdict=wrong; wrong=$((wrong+1)); fi
  echo "| $scenario | $arm | $(basename "$dir") | $validity | $truth | $status | ${suspects:-—} | $verdict |"
done < <(find "$root" -name run.json | sort)
echo
echo "Scored runs: correct $correct, partial $partial, wrong $wrong, abstain $abstain. Unscored intervention runs: $unscored."
