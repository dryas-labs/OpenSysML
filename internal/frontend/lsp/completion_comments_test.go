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

func TestCompletionSuppressesCommentBodies(t *testing.T) {
	for _, tc := range []struct{ name, text string }{
		{"documentation", "package P { part def Battery { doc /* this is a basic Battery th| */ } }"},
		{"unfinished-documentation", "package P { part def Battery { doc /* this is a basic Battery th|"},
		{"documentation-opener", "package P { doc /*|*/ }"},
		{"documentation-end", "package P { doc /* text |*/ }"},
		{"named-documentation", "package P { doc description locale \"en\" /* SI::| */ }"},
		{"regular-comment", "package P { /* SI::| */ }"},
		{"named-comment", "package P { comment note /* this| */ }"},
		{"block-note", "package P { //* this| */ }"},
		{"unfinished-block-note", "package P { //* this|"},
		{"line-note", "package P { // SI::|\n}"},
		{"line-note-eof", "// this|"},
		{"line-note-crlf", "// this|\r\npackage P;"},
		{"unicode-doc", "package P { doc /* 电池 🙂 th| */ }"},
		{"kerML-doc", "package P { class Battery { doc /* Battery th| */ } }"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file := "comment.sysml"
			if tc.name == "kerML-doc" {
				file = "comment.kerml"
			}
			items := completionAtCursor(t, tc.text, file)
			if len(items) != 0 {
				t.Fatalf("comment body offered %d code completions", len(items))
			}
		})
	}
}

func TestCompletionResumesOutsideCommentBodies(t *testing.T) {
	for _, text := range []string{
		"package P { part def Battery; /* note */| }",
		"package P { part def Battery; //* note */| }",
		"package P { part def Battery; // note\n| }",
		"package P { part def Battery; // note\r\n| }",
		"package P { part def Battery; |/* note */ }",
		"package P { part def Battery; doc /* note */ part b : Bat|; }",
		"package P { part def Battery; attribute text = \"/* not a comment\"; part b : Bat|; }",
		"package P { part def Battery; part def '/* not a comment'; part b : Bat|; }",
	} {
		items := completionAtCursor(t, text, "outside.sysml")
		found := false
		for _, item := range items {
			found = found || item.Label == "Battery"
		}
		if !found {
			t.Errorf("completion did not resume outside a comment: %q", text)
		}
	}
}

func completionAtCursor(t *testing.T, marked, file string) []protocol.CompletionItem {
	t.Helper()
	offset := strings.IndexByte(marked, '|')
	if offset < 0 {
		t.Fatal("fixture lacks cursor")
	}
	text := strings.Replace(marked, "|", "", 1)
	name := filepath.Join(t.TempDir(), file)
	ws := model.NewWorkspace()
	ws.Open(name, []byte(text), 1)
	server := NewServer(ws)
	result, err := server.Completion(context.Background(), &protocol.CompletionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: uri.File(name)},
			Position:     offsetToPosition([]byte(text), offset),
		},
	})
	if err != nil || result == nil {
		t.Fatalf("completion: %v", err)
	}
	if result.IsIncomplete {
		t.Fatal("completion must not request further results inside a comment")
	}
	return result.Items
}
