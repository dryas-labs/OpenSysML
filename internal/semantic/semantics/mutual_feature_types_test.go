package semantics

import "testing"

func TestMutualSpecializationDoesNotEraseFeatureTypes(t *testing.T) {
	m, root := buildModel(t, `part def A :> B; part def B :> A; part x : A, B;`)
	types := m.FeatureTypeSet(sym(t, root, "x"))
	if len(types) != 2 || !containsElement(types, sym(t, root, "A")) || !containsElement(types, sym(t, root, "B")) {
		t.Fatalf("equivalent type declarations must not eliminate one another: %v", types)
	}
	m, root = buildModel(t, `part def A; part def B :> A; part x : A, B;`)
	types = m.FeatureTypeSet(sym(t, root, "x"))
	if len(types) != 1 || !containsElement(types, sym(t, root, "B")) {
		t.Fatalf("a strictly more general type remains redundant: %v", types)
	}
}
