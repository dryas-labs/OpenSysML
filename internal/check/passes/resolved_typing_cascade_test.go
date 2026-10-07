package passes

import (
	"strings"
	"testing"
)

func TestResolvedTypingSurvivesOtherResolutionErrors(t *testing.T) {
	for _, tc := range []struct{ source, message string }{
		{`package P { part def Tank { port inlet : ~Tank; } part t : Tank; item x = t.inlet.absent; }`, f90PortMessage},
		{`package P { part def Tank; port inlet : Tank; item x = inlet.absent; }`, "A port must be typed by port definitions."},
	} {
		diags := libraryDiags(t, tc.source)
		found := false
		for _, d := range diags {
			if strings.Contains(d.Message, tc.message) {
				found = true
				if d.Span.Offset > strings.Index(tc.source, "item x") {
					t.Fatal("kind error anchored on the cascade")
				}
			}
		}
		if !found {
			t.Fatalf("resolved type error suppressed: %v", diags)
		}
	}
	for _, src := range []string{
		`package P { port p : Missing; }`,
		`package P { port def A; port p : ~A; part broken : Missing; }`,
	} {
		for _, d := range libraryDiags(t, src) {
			if d.Code == "type" || d.Code == "usage-typing" {
				t.Fatalf("type cascade: %v", d)
			}
		}
	}
}
