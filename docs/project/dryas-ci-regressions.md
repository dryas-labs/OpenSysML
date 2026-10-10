# Downstream CI regressions

**English** | [简体中文](dryas-ci-regressions.zh-Hans.md)

This record explains the fixes behind the broader PR checks, not a full-suite pass claim.
No benchmark thresholds, category maps, adjudications or downloaded corpus files are changed.

| Failure | Repair and preserved check |
| --- | --- |
| Missing pages in strict site build | Keep downstream engineering records in the repository, explicitly outside the published site, following the existing site policy. Strict checking stays enabled. |
| Node install fails | Regenerate missing platform-package lock entries from the published packages; retain dependency versions and validate with clean npm ci. |
| Self-model registry mismatch | Mark TypeCheckPass element-scoped in the self-model, matching its implementation. |
| Stale library snapshot | Regenerate the derived snapshot using the maintained resolver; compare it with a fresh source load. |
| Missing independent type errors and inconsistent incremental answers | Remove document-wide expression suppression after a name error; retain syntax gating and suppress expression findings only over unresolved spans. |
| Cross-feature type mismatch | Derive an owned cross feature's kind-implied types from its end, including Parts::Part for untyped part ends. |
| rad and min references fail in runtime fixtures | Select the SI unit explicitly instead of relying on import order against TrigFunctions::rad or QuantityCalculations::min. Expected numerical results and error messages are unchanged. |
| Rebinding fixture imports two different Car types | Place the fallback import in the enclosing namespace so the test exercises lexical shadowing. Original object-identity and rebinding assertions remain. |
| Positive filtered-import fixture also imports two different Good elements | Qualify the metadata reference in the positive fixture; keep the original collision as a separate negative test. The positive fixture must still reject only Bad. |
| REPL fixture types a part by a package | Reference an actual part definition inside that package. Preserve the root-import isolation assertion and quote the file path for portability. |
| SI dimension assertion depends on an ambiguous unit name | Assert the original two unresolved MagneticDipoleMomentUnit references separately. In dimensional controls only, qualify the same ISQElectromagnetism type the existing assertions describe. Keep all dimensional assertions and leave vendored library text unchanged. |

The native maintenance CI includes these regression groups. Broader upstream checks
remain required for assessing the whole change; passing selected groups does not
establish full conformance or make unrelated failures acceptable.
