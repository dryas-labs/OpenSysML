// Census: validatePartUsagePartDefinition
// Expect: clean
// Modified by DRYAS maintainers: preserve the valid implicit-Part model as a positive control.
// SysML 8.3.11.3 checkPartUsageSpecialization and KerML 8.3.3.3.4 deriveFeatureType
// supply Parts::Part alongside I. This is not a violating model.
package P { item def I; part p : I; }
