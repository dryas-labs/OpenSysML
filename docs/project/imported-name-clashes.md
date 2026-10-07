# Imported name clashes

Ordinary lookup hides imported memberships that give one name to distinct
elements, following [KerML 1.0 §7.2.5.4](https://www.omg.org/spec/KerML/1.0/PDF).
Repeated imports and aliases of one element remain valid. An independent outer
or global declaration can still supply the name. Root imports cannot be restored
by the global index, hidden re-exports do not interfere with an independent valid
import, and inherited imports are checked together rather than in import order.
Invocation overload candidate collection retains its separate behavior.

## Upstream PR corpus adjudication

The corpus is the `2026-08` release pinned by `scripts/pilot-pin.sh`, commit
`692170b71867353b8f90341e61556f49a5beb0e5`. On the upstream PR base, four previously clean files gain 32
diagnostics: 21 references to hidden imported names and 11 downstream diagnostics.
These movements expose ambiguous references previously bound by import order;
they are not 32 independent defects.

| Corpus file | Primary / downstream | Cause |
| --- | --- | --- |
| `kerml-examples/Simple Tests/Imports.kerml` | 3 / 0 | `S` imports distinct `P::A` and `Q::A`; recursive `Q::**` also surfaces distinct `Q::D` and `Q::Q1::D`. `S1` imports `P::A` and re-exported `Q::A` through `R`. References at lines 35, 36 and 44 have no unique imported binding. |
| `sysml-examples/Vehicle Example/Annex_A_VehicleViews.sysml` | 6 / 0 | The recursive import of `SimpleVehicleModel::VehicleConfigurations::VehicleConfiguration_b` surfaces distinct `vehicle_b` declarations in its parts and state models. References at lines 659, 670–672, 686 and 712 are hidden. |
| `sysml-examples/Vehicle Example/SysML v2 Spec Annex A SimpleVehicleModel.sysml` | 9 / 11 | The same recursive import makes `vehicle_b` ambiguous at lines 1094, 1106–1108, 1166, 1198, 1216, 1241 and 1472. The unresolved vehicle specialization causes eight engine/redefinition diagnostics and three action/port diagnostics. |
| `sysml-validation/13-Model Containment/13a-Model Containment.sysml` | 3 / 0 | Wildcard imports of `2a-Parts Interconnection` and `8-Requirements` surface distinct `vehicle1_c1`, `Engine` and `Transmission` declarations. References at lines 21, 56 and 57 are hidden. |

`TestPilotImportedNameClashQualification` loads each complete corpus batch, then
qualifies only those 21 primary references in memory. The vehicle controls select
`SimpleVehicleModel::VehicleConfigurations::VehicleConfiguration_b::PartsTree::vehicle_b`;
the containment controls select the corresponding declarations in `8-Requirements`.
The import example selects `Imports::P::A` and `Imports::Q::D`.
All four files then have no diagnostics, including the eleven downstream messages.
These are disambiguation controls, not a claim about which alternative the model
author intended. Downloaded models remain unchanged.

The upstream ratchet records the resulting 3, 6, 20 and 3 diagnostics. This adjudication
does not establish agreement with the Java Pilot validator or complete import-rule
conformance. The independent reference-validator comparison remains separate.

## Verification

```sh
go test ./internal/semantic/resolve ./internal/semantic/semantics ./internal/check/passes
OPENSYSML_REQUIRE_PILOT_CORPORA=1 go test ./tests/corpus -run 'TestPilotCorporaDiagnostics|TestPilotImportedNameClashQualification' -count=1
```

## DRYAS corpus expectations

The DRYAS fork also carries previously reviewed semantic fixes that the upstream
PR intentionally excludes. Its own expectation file must therefore not retain
those fixed false alarms. A comparison of the pre-integration maintenance commit
`4ede7ba08` and this integration found identical diagnostics for every file in the
same pinned corpus: 58 KerML examples, 99 SysML examples and 56 validation files.

| File | Upstream PR record | DRYAS record | Basis |
| --- | --- | --- | --- |
| Geometry Examples/SimpleQuadcopter.sysml | 2 | 0 | Existing kind-implied Part typing fix; SysML 2.0 7.11.2 and 8.4.7.2. |
| Simple Tests/IndividualTest.sysml | 1 | 0 | Same existing Part typing correction. |
| Simple Tests/ItemTest.sysml | 1 | 0 | Same existing Part typing correction. |
| Simple Tests/PartTest.sysml | 4 | 0 | Existing correction for Base-reachable specialization/subsetting cycles; KerML 1.0 7.3.2.3. Closed cycles without a Base path remain rejected by a separate test. |
| Vehicle Example/SysML v2 Spec Annex A SimpleVehicleModel.sysml | 20 | 18 | Eighteen identical messages were observed before and after integration. This records the maintained branch's existing behavior, not a new claim of improvement from this backport. |

The four import-clash files retain all 21 primary unresolved-reference anchors.
Qualifying those references clears every diagnostic in the four files, as checked
by `TestPilotImportedNameClashQualification`. The DRYAS totals for those files are
3, 6, 18 and 3. No golden-generation switch was used; only the five documented
branch differences were edited. The exact-count test still rejects any other
increase or decrease. Historical M0 inputs, labels, thresholds and scores are
untouched; this is the engine repository's separate Go regression baseline.
