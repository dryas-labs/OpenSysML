// SPDX-License-Identifier: Apache-2.0
package lsp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"

	"github.com/Open-MBEE/OpenSysML/internal/workspace/libs"
	"github.com/Open-MBEE/OpenSysML/internal/workspace/model"
)

func lazyCompletionServer(t testing.TB, markdown bool) *Server {
	t.Helper()
	s := NewServer(model.NewWorkspace())
	format := protocol.PlainText
	if markdown {
		format = protocol.Markdown
	}
	result, err := s.Initialize(context.Background(), &protocol.InitializeParams{Capabilities: protocol.ClientCapabilities{
		TextDocument: &protocol.TextDocumentClientCapabilities{Completion: &protocol.CompletionTextDocumentClientCapabilities{
			CompletionItem: &protocol.CompletionTextDocumentClientCapabilitiesItem{
				DocumentationFormat: []protocol.MarkupKind{format},
				ResolveSupport:      &protocol.CompletionTextDocumentClientCapabilitiesItemResolveSupport{Properties: []string{"documentation"}},
			},
		}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Capabilities.CompletionProvider.ResolveProvider {
		t.Fatal("completion resolve is not advertised")
	}
	return s
}

func resolvedCandidate(t testing.TB, s *Server, name, src, marker, label string) protocol.CompletionItem {
	t.Helper()
	offset := strings.Index(src, marker) + len(marker)
	if offset < len(marker) {
		t.Fatal("fixture lacks completion position")
	}
	result, err := s.Completion(context.Background(), &protocol.CompletionParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri.File(name)}, Position: offsetToPosition([]byte(src), offset),
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range result.Items {
		if item.Label == label {
			if item.Documentation != nil {
				t.Fatal("initial completion eagerly loaded documentation")
			}
			if item.Data == nil {
				t.Fatal("candidate has no resolve handle")
			}
			// Exercise the JSON wire representation, not only a Go-only data type.
			encoded, err := json.Marshal(item)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), filepath.Dir(name)) {
				t.Fatal("handle contains a host path")
			}
			var wire protocol.CompletionItem
			if err := json.Unmarshal(encoded, &wire); err != nil {
				t.Fatal(err)
			}
			return wire
		}
	}
	t.Fatalf("missing candidate %q", label)
	return protocol.CompletionItem{}
}

func resolveProse(t testing.TB, s *Server, item protocol.CompletionItem) string {
	t.Helper()
	result, err := s.CompletionResolve(context.Background(), &item)
	if err != nil {
		t.Fatal(err)
	}
	doc, ok := result.Documentation.(protocol.MarkupContent)
	if !ok {
		t.Fatalf("documentation = %#v", result.Documentation)
	}
	copy := *result
	copy.Documentation = item.Documentation
	if !reflect.DeepEqual(copy, item) {
		t.Fatal("resolve changed insertion, sorting or other original candidate fields")
	}
	return doc.Value
}

func TestCompletionResolveBundledDocumentation(t *testing.T) {
	s := lazyCompletionServer(t, true)
	name := filepath.Join(t.TempDir(), "voltage.sysml")
	src := "package P { attribute v :> ISQ::vo; }"
	s.ws.Open(name, []byte(src), 1)
	item := resolvedCandidate(t, s, name, src, "ISQ::vo", "voltage")
	prose := resolveProse(t, s, item)
	for _, want := range []string{"attribute voltage: ElectricPotentialDifferenceValue :> scalarQuantities", "ISQ::voltage", "ISQElectromagnetism::voltage", "ISQElectromagnetism.sysml", "source: item 6-11.3", "remarks: For an electric field", "121-11-27."} {
		if !strings.Contains(prose, want) {
			t.Errorf("resolved documentation missing %q", want)
		}
	}
}

func TestCompletionResolveKeepsAliasAndShortNameInsertion(t *testing.T) {
	s := lazyCompletionServer(t, true)
	root := t.TempDir()
	s.ws.SetOnDisk(filepath.Join(root, "definitions.sysml"), []byte("package Parts { part def <Bat> Battery { doc /* Battery documentation. */ } alias Cell for Battery; part def Plain; }"))
	name := filepath.Join(root, "system.sysml")
	src := "package P { private import Parts::*; part b : Ba; }"
	s.ws.Open(name, []byte(src), 1)
	for _, label := range []string{"Battery", "Cell", "Bat"} {
		item := resolvedCandidate(t, s, name, src, "b : Ba", label)
		prose := resolveProse(t, s, item)
		for _, want := range []string{"Parts::Battery", "definitions.sysml", "Battery documentation.", label} {
			if !strings.Contains(prose, want) {
				t.Errorf("%s missing %q", label, want)
			}
		}
	}
	plain := resolvedCandidate(t, s, name, src, "b : Ba", "Plain")
	if prose := resolveProse(t, s, plain); !strings.Contains(prose, "part def Plain") || strings.Contains(prose, "Battery documentation.") {
		t.Errorf("undocumented candidate = %q", prose)
	}
}

func TestCompletionResolveDistinguishesSameNamedDeclarations(t *testing.T) {
	s := lazyCompletionServer(t, true)
	root := t.TempDir()
	s.ws.SetOnDisk(filepath.Join(root, "types.sysml"), []byte("package Left { part def Battery { doc /* LEFT DOC */ } } package Right { part def Battery { doc /* RIGHT DOC */ } }"))
	name := filepath.Join(root, "system.sysml")
	src := "package P { part a : Left::Ba; part b : Right::Ba; }"
	s.ws.Open(name, []byte(src), 1)
	left := resolvedCandidate(t, s, name, src, "Left::Ba", "Battery")
	right := resolvedCandidate(t, s, name, src, "Right::Ba", "Battery")
	if p := resolveProse(t, s, right); !strings.Contains(p, "RIGHT DOC") || strings.Contains(p, "LEFT DOC") {
		t.Fatalf("right = %q", p)
	}
	if p := resolveProse(t, s, left); !strings.Contains(p, "LEFT DOC") || strings.Contains(p, "RIGHT DOC") {
		t.Fatalf("left = %q", p)
	}
}

func TestCompletionResolveAllowsFurtherPrefixTypingButRejectsChangedDeclaration(t *testing.T) {
	s := lazyCompletionServer(t, true)
	root := t.TempDir()
	definition := filepath.Join(root, "definitions.sysml")
	s.ws.SetOnDisk(definition, []byte("package Parts { part def Battery { doc /* Original documentation. */ } }"))
	name := filepath.Join(root, "system.sysml")
	src := "package P { part b : Parts::Ba; }"
	s.ws.Open(name, []byte(src), 1)
	item := resolvedCandidate(t, s, name, src, "Parts::Ba", "Battery")
	s.ws.Update(name, []byte(strings.Replace(src, "Parts::Ba", "Parts::Bat", 1)), 2)
	if !strings.Contains(resolveProse(t, s, item), "Original documentation.") {
		t.Fatal("typing dropped valid candidate docs")
	}
	s.ws.SetOnDisk(definition, []byte("package Parts { part def Battery { doc /* Changed documentation. */ } }"))
	if _, err := s.CompletionResolve(context.Background(), &item); err == nil {
		t.Fatal("changed declaration accepted a stale completion handle")
	}
}

func TestCompletionResolvePlainTextCancellationAndInvalidHandles(t *testing.T) {
	s := lazyCompletionServer(t, false)
	name := filepath.Join(t.TempDir(), "plain.sysml")
	src := "package P { part def Battery { doc /* Complete plain text. */ } part b : Ba; }"
	s.ws.Open(name, []byte(src), 1)
	item := resolvedCandidate(t, s, name, src, "b : Ba", "Battery")
	text := resolveProse(t, s, item)
	if !strings.Contains(text, "Complete plain text.") || strings.Contains(text, "```") {
		t.Fatalf("plain text = %q", text)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.CompletionResolve(ctx, &item); err == nil {
		t.Fatal("cancelled resolve succeeded")
	}
	for _, changed := range []protocol.CompletionItem{
		{Label: "Other", Data: item.Data},
		{Label: "Battery", Data: map[string]any{"token": "unknown", "index": -1}},
	} {
		if _, err := s.CompletionResolve(context.Background(), &changed); err == nil {
			t.Fatal("invalid handle accepted")
		}
	}
	other := lazyCompletionServer(t, true)
	if _, err := other.CompletionResolve(context.Background(), &item); err == nil {
		t.Fatal("another session accepted the handle")
	}
	for range 10 {
		resolvedCandidate(t, s, name, src, "b : Ba", "Battery")
	}
	if _, err := s.CompletionResolve(context.Background(), &item); err == nil {
		t.Fatal("old completion batch was never evicted")
	}
	keyword := protocol.CompletionItem{Label: "part", Kind: protocol.CompletionItemKindKeyword}
	result, err := s.CompletionResolve(context.Background(), &keyword)
	if err != nil || !reflect.DeepEqual(*result, keyword) {
		t.Fatal("keyword without resolve data should be unchanged")
	}
}

func TestCompletionResolveReadsRecordedSourceWithoutReplacingIndex(t *testing.T) {
	root := t.TempDir()
	name := filepath.Join(root, "definitions.sysml")
	text := "package Parts { part def Battery { doc /* Recorded source documentation. */ } part def Other { doc /* Other documentation. */ } }"
	cache, err := libs.NewCacheIn(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	inputs := []model.Input{{Name: name, Content: []byte(text), Version: 1}}
	cold := model.NewWorkspace(model.WithRecordCache(cache))
	cold.OpenAll(inputs)
	cold.DiagnosticsAll([]string{name})
	ws := model.NewWorkspace(model.WithRecordCache(cache))
	ws.OpenAll(inputs)
	if !ws.Recorded(name) {
		t.Fatal("fixture did not load as a record")
	}
	s := lazyCompletionServer(t, true)
	s.ws = ws
	consumer := filepath.Join(root, "system.sysml")
	src := "package System { part b : Parts::Ba; }"
	ws.Open(consumer, []byte(src), 1)
	generation := ws.Generation()
	item := resolvedCandidate(t, s, consumer, src, "Parts::Ba", "Battery")
	if !ws.Recorded(name) {
		t.Fatal("completion enumeration hydrated the closed source")
	}
	prose := resolveProse(t, s, item)
	if !strings.Contains(prose, "Recorded source documentation.") || strings.Contains(prose, "Other documentation.") {
		t.Fatalf("wrong recorded doc: %q", prose)
	}
	if !ws.Recorded(name) || ws.Generation() != generation {
		t.Fatal("resolving documentation mutated the model index")
	}
}

func TestCompletionResolveSurvivesTypingAfterSameFileDeclaration(t *testing.T) {
	s := lazyCompletionServer(t, true)
	name := filepath.Join(t.TempDir(), "same-file.sysml")
	src := "package P { part def Battery { doc /* Same-file documentation. */ } part b : Ba; }"
	s.ws.Open(name, []byte(src), 1)
	item := resolvedCandidate(t, s, name, src, "b : Ba", "Battery")
	s.ws.Update(name, []byte(strings.Replace(src, "b : Ba", "b : Bat", 1)), 2)
	if !strings.Contains(resolveProse(t, s, item), "Same-file documentation.") {
		t.Fatal("prefix typing lost a stable declaration")
	}
}

func TestCompletionResolveRejectsChangedContext(t *testing.T) {
	for _, change := range []string{"import", "other-file", "new-file"} {
		t.Run(change, func(t *testing.T) {
			s := lazyCompletionServer(t, true)
			root := t.TempDir()
			defs := filepath.Join(root, "definitions.sysml")
			s.ws.SetOnDisk(defs, []byte("package Parts { part def Battery; } package Other { part def Battery; }"))
			name := filepath.Join(root, "system.sysml")
			src := "package P { private import Parts::*; part b : Ba; }"
			s.ws.Open(name, []byte(src), 1)
			item := resolvedCandidate(t, s, name, src, "b : Ba", "Battery")
			switch change {
			case "import":
				s.ws.Update(name, []byte(strings.Replace(src, "Parts::*", "Other::*", 1)), 2)
			case "other-file":
				s.ws.SetOnDisk(defs, []byte("package Parts { part def Battery; } package Other { part def Different; }"))
			case "new-file":
				s.ws.SetOnDisk(filepath.Join(root, "new.sysml"), []byte("package New;"))
			}
			if _, err := s.CompletionResolve(context.Background(), &item); err == nil {
				t.Fatal("changed context accepted old candidate documentation")
			}
		})
	}
}

func TestCompletionResolveSurvivesTypingBeforeSameFileDeclaration(t *testing.T) {
	s := lazyCompletionServer(t, true)
	name := filepath.Join(t.TempDir(), "forward.sysml")
	src := "package P { part b : Ba; part def Battery { doc /* Forward documentation. */ } }"
	s.ws.Open(name, []byte(src), 1)
	item := resolvedCandidate(t, s, name, src, "b : Ba", "Battery")
	s.ws.Update(name, []byte(strings.Replace(src, "b : Ba", "b : Bat", 1)), 2)
	if !strings.Contains(resolveProse(t, s, item), "Forward documentation.") {
		t.Fatal("continued typing lost forward declaration documentation")
	}
}

// This measures the native request handlers separately; it is not an editor
// latency guarantee, nor a cold process-start benchmark.
func BenchmarkCompletionDocumentation(b *testing.B) {
	s := lazyCompletionServer(b, true)
	name := filepath.Join(b.TempDir(), "voltage.sysml")
	src := "package P { attribute v :> ISQ::vo; }"
	s.ws.Open(name, []byte(src), 1)
	params := &protocol.CompletionParams{TextDocumentPositionParams: protocol.TextDocumentPositionParams{
		TextDocument: protocol.TextDocumentIdentifier{URI: uri.File(name)},
		Position:     offsetToPosition([]byte(src), strings.Index(src, "ISQ::vo")+len("ISQ::vo")),
	}}
	b.Run("list", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := s.Completion(context.Background(), params); err != nil {
				b.Fatal(err)
			}
		}
	})
	item := resolvedCandidate(b, s, name, src, "ISQ::vo", "voltage")
	resolveProse(b, s, item)
	b.Run("resolve-warm", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if _, err := s.CompletionResolve(context.Background(), &item); err != nil {
				b.Fatal(err)
			}
		}
	})
}
