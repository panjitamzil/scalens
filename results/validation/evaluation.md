## cpu

| Scenario | Arm | Run | Validity | Injected fault | Status | Suspects | Verdict |
|---|---|---|---|---|---|---|---|
| cpu | control | 20261116T070816Z-2 | contended | cpu_saturation | suspected | cpu_saturation | correct |
| cpu | control | 20261116T071409Z-3 | contended | cpu_saturation | suspected | cpu_saturation, slow_database_operation | partial (extra suspect) |
| cpu | control | 20261116T152725Z-1 | contended | cpu_saturation | suspected | cpu_saturation | correct |
| cpu | hpa | 20261116T065843Z-1 | contended | unscored | insufficient_evidence | — | unscored |
| cpu | hpa | 20261116T070517Z-2 | contended | unscored | suspected | cpu_saturation | unscored |
| cpu | hpa | 20261116T071115Z-3 | contended | unscored | insufficient_evidence | — | unscored |
| cpu | scale | 20261116T065400Z-1 | contended | unscored | suspected | cpu_saturation | unscored |
| cpu | scale | 20261116T070221Z-2 | contended | unscored | insufficient_evidence | — | unscored |
| cpu | scale | 20261116T153024Z-3 | contended | unscored | no_anomaly | — | unscored |

Scored runs: correct 2, partial 1, wrong 0, abstain 0. Unscored intervention runs: 6.

## cpu-v2

| Scenario | Arm | Run | Validity | Injected fault | Status | Suspects | Verdict |
|---|---|---|---|---|---|---|---|
| cpu | control | 20261117T094259Z-1 | contended | cpu_saturation | suspected | cpu_saturation | correct |
| cpu | control | 20261117T095655Z-2 | contended | cpu_saturation | suspected | cpu_saturation | correct |
| cpu | control | 20261117T100234Z-3 | contended | cpu_saturation | suspected | cpu_saturation | correct |
| cpu | hpa | 20261117T094838Z-1 | contended | unscored | insufficient_evidence | — | unscored |
| cpu | hpa | 20261117T095409Z-2 | contended | unscored | insufficient_evidence | — | unscored |
| cpu | hpa | 20261117T095948Z-3 | contended | unscored | insufficient_evidence | — | unscored |
| cpu | scale | 20261117T094550Z-1 | contended | unscored | no_anomaly | — | unscored |
| cpu | scale | 20261117T095123Z-2 | contended | unscored | no_anomaly | — | unscored |
| cpu | scale | 20261117T100526Z-3 | contended | unscored | no_anomaly | — | unscored |

Scored runs: correct 3, partial 0, wrong 0, abstain 0. Unscored intervention runs: 6.

## database

| Scenario | Arm | Run | Validity | Injected fault | Status | Suspects | Verdict |
|---|---|---|---|---|---|---|---|
| database | control | 20261116T153620Z-1 | contended | slow_database_operation | suspected | slow_database_operation | correct |
| database | control | 20261116T154859Z-2 | contended | slow_database_operation | suspected | slow_database_operation | correct |
| database | control | 20261116T161630Z-3 | contended | slow_database_operation | suspected | slow_database_operation | correct |
| database | healthy | 20261116T153316Z-1 | contended | none | no_anomaly | — | correct |
| database | healthy | 20261116T160136Z-2 | contended | none | no_anomaly | — | correct |
| database | healthy | 20261116T161339Z-3 | contended | none | no_anomaly | — | correct |
| database | hpa | 20261116T154549Z-1 | contended | slow_database_operation | suspected | slow_database_operation | correct |
| database | hpa | 20261116T155828Z-2 | contended | slow_database_operation | suspected | slow_database_operation | correct |
| database | hpa | 20261116T161033Z-3 | contended | slow_database_operation | suspected | slow_database_operation | correct |
| database | index | 20261116T154251Z-1 | contended | none | no_anomaly | — | correct |
| database | index | 20261116T155534Z-2 | contended | none | no_anomaly | — | correct |
| database | index | 20261116T160740Z-3 | contended | none | no_anomaly | — | correct |
| database | scale | 20261116T153934Z-1 | contended | slow_database_operation | suspected | slow_database_operation | correct |
| database | scale | 20261116T155208Z-2 | contended | slow_database_operation | suspected | slow_database_operation | correct |
| database | scale | 20261116T160426Z-3 | contended | slow_database_operation | suspected | slow_database_operation | correct |

Scored runs: correct 15, partial 0, wrong 0, abstain 0. Unscored intervention runs: 0.

