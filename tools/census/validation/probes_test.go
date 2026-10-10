// Modified by DRYAS maintainers: distinguish construction-satisfied constraints from reported violations.
package validation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Open-MBEE/OpenSysML/internal/workspace/model"
	"github.com/Open-MBEE/OpenSysML/tools/oracle/repo"
)

// TestProbesReportTheirConstraint is the evidence behind every ✅/⚠️ row of the
// census: violating models require their named diagnostic; construction
// controls must remain completely diagnostic-free.
func TestProbesReportTheirConstraint(t *testing.T) {
	root, err := repo.Root()
	if err != nil {
		t.Fatal(err)
	}
	probes, err := loadProbes(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(probes) == 0 {
		t.Fatalf("no probes under %s", probesDir)
	}
	for _, p := range probes {
		t.Run(filepath.Base(p.Path), func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(root, p.Path))
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Base(p.Path)
			ws := model.NewWorkspace()
			ws.Open(name, content, 1)
			findings := ws.Diagnostics(name)
			if p.Clean {
				if len(findings) != 0 {
					t.Fatalf("%s: construction control must be clean, got %v", p.Path, findings)
				}
				return
			}
			var seen []string
			for _, d := range findings {
				line := d.Severity.String() + ": " + d.Message
				if d.Severity.String() == p.Severity && strings.Contains(d.Message, p.Message) {
					return
				}
				seen = append(seen, line)
			}
			t.Errorf("%s: expected a %s containing %q for %s; got:\n  %s",
				p.Path, p.Severity, p.Message, p.Constraint, strings.Join(seen, "\n  "))
		})
	}
}
