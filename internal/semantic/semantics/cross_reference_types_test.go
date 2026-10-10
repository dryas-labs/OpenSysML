// SPDX-License-Identifier: Apache-2.0
// Added by DRYAS maintainers: cross-feature kind-implied type regression.
package semantics

import (
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"testing"
)

func TestCrossReferenceRetainsUntypedPartBase(t *testing.T) {
	m, root := buildModelWithStdlib(t, `package P {
     part def TireBead;
     connection def Seat { end [1] part bead : TireBead; end [1] part rim; }
 }`)
	for _, name := range []string{"bead", "rim"} {
		end := nestedSym(t, root, "P::Seat::"+name)
		cross := m.CrossFeature(end)
		if cross == nil {
			t.Fatalf("%s has no owned cross feature", name)
		}
		a, b := m.FeatureTypeSet(cross), m.FeatureTypeSet(end)
		if len(a) == 0 || len(a) != len(b) {
			t.Fatalf("%s: cross has %d types, end has %d", name, len(a), len(b))
		}
		for _, want := range b {
			found := false
			for _, got := range a {
				if symbols.SameElement(got, want) {
					found = true
				}
			}
			if !found {
				t.Errorf("%s: cross missing end type %s", name, want.Name)
			}
		}
	}
}
