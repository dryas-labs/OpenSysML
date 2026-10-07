package passes

import (
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/parser"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
	"testing"
)

func TestCyclesNeedAnActualBasePath(t *testing.T) {
	const src = `classifier A :> B; classifier B :> A;`
	file := source.New("closed.kerml", []byte(src))
	root := parser.New(file).ParseFile()
	idx := symbols.NewIndex()
	idx.AddDocument(file.Name(), root)
	if !hasCode(Analyze(file.Name(), root, nil, idx), "specialization-cycle") {
		t.Fatal("an entirely closed graph without Anything must still be reported")
	}
	for _, d := range w8cLibraryDiagnostics(t, "with-base.sysml", `part p1 :> p2; part p2 :> p1; part p3 :> p3;`) {
		if d.Code == "specialization-cycle" {
			t.Fatalf("valid base-reachable cycle rejected: %v", d)
		}
	}
}
