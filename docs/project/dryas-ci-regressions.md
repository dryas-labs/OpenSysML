# Downstream CI regressions

**English** | [简体中文](dryas-ci-regressions.zh-Hans.md)

This record explains the fixes behind the broader PR checks, not a full-suite pass claim.
Historical M0 thresholds, category maps, adjudications and scores are unchanged. Downloaded corpus files are unmodified. The engine repository's census correction and new advisory reference run are recorded separately below.

| Failure | Repair and preserved check |
| --- | --- |
| Missing pages in strict site build | Keep downstream engineering records in the repository, explicitly outside the published site, following the existing site policy. Strict checking stays enabled. |
| Node install fails | Regenerate missing platform-package lock entries from the published packages; retain dependency versions and validate with clean npm ci. |
| Self-model registry mismatch | Mark TypeCheckPass element-scoped in the self-model, matching its implementation. |
| Stale library snapshot | Regenerate the derived snapshot using the maintained resolver; compare it with a fresh source load. |
| Missing independent type errors and inconsistent incremental answers | Remove document-wide expression suppression after a name error; retain syntax gating and suppress expression findings only over unresolved spans. |
| Cross-feature type mismatch | Derive an owned cross feature's kind-implied types from its end, including Parts::Part for untyped part ends. |
| rad and min references fail in runtime fixtures | Select the SI unit explicitly instead of relying on import order against TrigFunctions::rad or QuantityCalculations::min. Expected numerical results and error categories are unchanged; diagnostic source echoes use the explicitly qualified unit spelling. |
| Negative sequence-index fixtures counted as clean examples | Require their exact static errors, consistent with their existing runtime-error expectations; check both fixtures are visited and keep all other examples diagnostic-free. |
| Rebinding fixture imports two different Car types | Place the fallback import in the enclosing namespace so the test exercises lexical shadowing. Original object-identity and rebinding assertions remain. |
| Positive filtered-import fixture also imports two different Good elements | Qualify the metadata reference in the positive fixture; keep the original collision as a separate negative test. The positive fixture must still reject only Bad. |
| REPL fixture types a part by a package | Reference an actual part definition inside that package. Preserve the root-import isolation assertion and quote the file path for portability. |
| SI dimension assertion depends on an ambiguous unit name | Assert the original two unresolved MagneticDipoleMomentUnit references separately. In dimensional controls only, qualify the same ISQElectromagnetism type the existing assertions describe. Keep all dimensional assertions and leave vendored library text unchanged. |
| Part-definition census probe expects an error from a valid part | Retain the model as a clean control. SysML §8.3.11.3 and KerML §8.3.3.3.4 supply Parts::Part implicitly; classify it separately as satisfied by construction rather than count it as observed rejection coverage. |

The native maintenance CI includes these regression groups. Broader upstream checks
remain required for assessing the whole change; passing selected groups does not
establish full conformance or make unrelated failures acceptable.

## Downstream reference record

The retained upstream [record](pilot-differential-baseline.json) is unchanged.
[dryas-pilot-differential-baseline.json](dryas-pilot-differential-baseline.json) records
an actual downstream run of all 380 files with Pilot 2026-08 / 0.62.0, the same
reference bridges, category maps, standard library and corpus versions. The
provenance guard checks this current record. The self-model correction changes an
input value but produces no diagnostic movement in the 45-file examples root.

The launcher label omits the native `.exe` suffix so an unchanged run can reproduce
on either platform; the actual invocation keeps the complete filename. Paths inside diagnostic records
are normalized to repository-relative forward-slash paths before attribution.
Unattributed model diagnostics are checked separately from JVM warning output.
This record is advisory, separate from M0 scores and engine-selection gates.

| Comparison measure | Upstream historical run | Downstream run |
| --- | ---: | ---: |
| Files | 380 | 380 |
| Fully agreeing files | 345 | 348 |
| Agreed diagnostics | 38 | 38 |
| Severity mismatches | 3 | 3 |
| OpenSysML-only diagnostics | 41 | 55 |
| Pilot-only diagnostics | 1614 | 1614 |
| Pilot diagnostics in total | 1655 | 1655 |

Every changed diagnostic group has the following explanation. More or fewer
OpenSysML-only findings are not, by themselves, evidence of worse or better accuracy.

| Input and lines | Change from upstream history | Basis |
| --- | --- | --- |
| Geometry Examples/SimpleQuadcopter.sysml:14,114; Simple Tests/IndividualTest.sysml:12; ItemTest.sysml:11; PartTest.sysml:51,52,53,55 | Remove eight part-typing findings | The maintained kind-implied typing retains Parts::Part; the corresponding part and cross-feature regression tests require acceptance. |
| tests/testdata/passes/constraints.sysml:2,3; specialization-cycle-pair.sysml:4,5; specialization-cycle-self.sysml:4; specialization-cycle-three.sysml:4,5,6 | Remove eight cycle findings | The maintained resolver accepts legal specialization cycles and retains their implicit bases. The multiplicity finding at constraints.sysml:9 remains. |
| Vehicle Example/Annex_A_VehicleViews.sysml:659,670,671,672,686,712 | Add six unresolved imported-name findings | The maintained imported-name clash behavior; no generic unresolved message is remapped to a kind error. |
| Vehicle Example/SysML v2 Spec Annex A SimpleVehicleModel.sysml:1094,1106,1107,1108,1166,1169,1172,1173,1174,1178,1181,1182,1183,1198,1216,1241,1472,1474 | Add eighteen unresolved imported-name findings | Same imported-name behavior; source files and expected corpus gate remain unchanged. |
| 13-Model Containment/13a-Model Containment.sysml:21,56,57; Simple Tests/Imports.kerml:35,36,44 | Add six unresolved imported-name findings | Same imported-name behavior, including the KerML control. See [Imported-name clashes](imported-name-clashes.md). |

The net change is 16 removed findings and 30 added imported-name findings. All
100 training files remain clean on both sides. These are explanations of this
reference comparison, not new human adjudications or a general conformance claim.
