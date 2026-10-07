package lsp

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"github.com/Open-MBEE/OpenSysML/internal/workspace/model"
)

func TestHoverIncludesOwnedDocumentation(t *testing.T) {
	for _, tc := range []struct{ name, text, at, want string }{
		{"definition", "package P { part def Battery { doc /* Battery owner documentation. */ } }", "Battery", "Battery owner documentation."},
		{"usage", "package P { part battery { doc /* Battery usage documentation. */ } }", "battery", "Battery usage documentation."},
		{"package", "package Pack { doc /* Package documentation. */ }", "Pack", "Package documentation."},
		{"kerML-class", "package P { class Battery { doc /* Class documentation. */ } }", "Battery", "Class documentation."},
		{"named-doc", "package P { part def Battery { doc description locale \"en\" /* Named documentation. */ } }", "Battery", "Named documentation."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ws := model.NewWorkspace()
			s := NewServer(ws)
			initMarkdownHover(t, s)
			ext := ".sysml"
			if tc.name == "kerML-class" {
				ext = ".kerml"
			}
			name := filepath.Join(t.TempDir(), "owned"+ext)
			ws.Open(name, []byte(tc.text), 1)
			hover := hoverInSrc(t, s, name, tc.text, strings.Index(tc.text, tc.at))
			if !strings.Contains(hover.Contents.Value, tc.want) {
				t.Fatalf("missing owned doc: %q", hover.Contents.Value)
			}
			if strings.Contains(hover.Contents.Value, "/*") {
				t.Fatal("Markdown hover includes a comment delimiter")
			}
		})
	}
}

func TestHoverOwnedDocumentationPreservesFullTextAndOwnership(t *testing.T) {
	ws := model.NewWorkspace()
	s := NewServer(ws)
	initMarkdownHover(t, s)
	name := filepath.Join(t.TempDir(), "owners.sysml")
	src := "package P {\n// Leading summary.\npart def Battery {\n" +
		"doc /*\n * First owned paragraph.\n" + strings.Repeat(" * A line of documentation.\n", 20) +
		" * **Final owned paragraph.**\n */\n" +
		"doc extra /* Second owned document. */\n" +
		"part child { doc /* Child-only documentation. */ }\n" +
		"}\npart def Other { doc /* Sibling-only documentation. */ }\n}"
	ws.Open(name, []byte(src), 1)
	hover := hoverInSrc(t, s, name, src, strings.Index(src, "Battery"))
	for _, want := range []string{"Leading summary.", "First owned paragraph.", "**Final owned paragraph.**", "Second owned document."} {
		if !strings.Contains(hover.Contents.Value, want) {
			t.Errorf("missing %q", want)
		}
	}
	for _, absent := range []string{"Child-only documentation.", "Sibling-only documentation."} {
		if strings.Contains(hover.Contents.Value, absent) {
			t.Errorf("leaked another declaration's doc: %q", absent)
		}
	}
	if got := strings.Count(hover.Contents.Value, "A line of documentation."); got != 20 {
		t.Errorf("doc lines = %d, want 20", got)
	}
}

func TestHoverReferencedOwnedDocumentationAcrossFiles(t *testing.T) {
	for _, markdown := range []bool{false, true} {
		ws := model.NewWorkspace()
		s := NewServer(ws)
		if markdown {
			initMarkdownHover(t, s)
		}
		root := t.TempDir()
		ws.SetOnDisk(filepath.Join(root, "definitions.sysml"), []byte("package Components { part def Battery { doc /* Full cross-file documentation. */ } }"))
		name := filepath.Join(root, "system.sysml")
		src := "package System { part battery : Components::Battery; }"
		ws.Open(name, []byte(src), 1)
		hover := hoverInSrc(t, s, name, src, strings.Index(src, "::Battery")+2)
		if !strings.Contains(hover.Contents.Value, "Full cross-file documentation.") {
			t.Errorf("missing cross-file doc: %q", hover.Contents.Value)
		}
	}
}

func TestHoverIncludesCompleteBundledVoltageDocumentation(t *testing.T) {
	s := NewServer(model.NewWorkspace())
	initMarkdownHover(t, s)
	name := filepath.Join(t.TempDir(), "voltage.sysml")
	src := "package P { private import SI::*; attribute outputVoltage redefines ISQ::voltage = 48[V]; }"
	s.ws.Open(name, []byte(src), 1)
	offset := strings.Index(src, "ISQ::voltage") + len("ISQ::")
	hover := hoverInSrc(t, s, name, src, offset)
	for _, want := range []string{"attribute voltage", "source: item 6-11.3", "measurement unit(s): V", "remarks: For an electric field", "121-11-27."} {
		if !strings.Contains(hover.Contents.Value, want) {
			t.Errorf("bundled hover omits %q", want)
		}
	}
	locations, err := s.Definition(context.Background(), &protocol.DefinitionParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri.File(name)}, Position: offsetToPosition([]byte(src), offset),
	}})
	if err != nil || len(locations) != 1 {
		t.Fatalf("definition: %v", err)
	}
	if locations[0].Range.End.Line-locations[0].Range.Start.Line < 10 {
		t.Fatal("definition range must still cover the full declaration")
	}
}
