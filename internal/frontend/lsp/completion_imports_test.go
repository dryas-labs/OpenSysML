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

func TestCompletionVisibleImportedNames(t *testing.T) {
	for _, tc := range []struct {
		name, definitions, consumer string
		want, absent                []string
	}{
		{"package-wildcard", `package Components { part def Battery; }`, `package System { private import Components::*; part battery : Battery; part secondaryBattery : Ba; }`, []string{"Battery"}, nil},
		{"unfinished-declaration", `package Components { part def Battery; }`, "package System { private import Components::*; part secondaryBattery : Ba\n}", []string{"Battery"}, nil},
		{"recursive-import", `package Components { package Cells { part def Battery; } }`, `package System { private import Components::**; part secondaryBattery : Ba; }`, []string{"Battery"}, nil},
		{"cyclic-reexport", `package Components { part def Battery; public import Library::*; } package Library { public import Components::*; }`, `package System { private import Library::*; part secondaryBattery : Ba; }`, []string{"Battery"}, nil},
		{"owned-clash-control", `package Left { part def Battery; } package Right { part def Battery; }`, `package System { private import Left::*; private import Right::*; item def Battery; part secondaryBattery : Ba; }`, []string{"Battery"}, nil},
		{"outer-clash-control", `package Left { part def Battery; } package Right { part def Battery; } item def Battery;`, `package System { private import Left::*; private import Right::*; part secondaryBattery : Ba; }`, []string{"Battery"}, nil},
		{"membership", `package Components { part def Battery; part def Motor; }`, `package System { private import Components::Battery; part secondaryBattery : Ba; }`, []string{"Battery"}, []string{"Motor"}},
		{"alias", `package Components { part def Battery; alias Cell for Battery; }`, `package System { private import Components::Cell; part secondaryBattery : Ba; }`, []string{"Cell"}, []string{"Battery"}},
		{"short-name", `package Components { part def <Bat> Battery; }`, `package System { private import Components::Bat; part secondaryBattery : Ba; }`, []string{"Bat", "Battery"}, nil},
		{"nested", `package Components { part def Battery; }`, `package System { private import Components::*; part def Assembly { part secondaryBattery : Ba; } }`, []string{"Battery"}, nil},
		{"private-member", `package Components { private part def Battery; public part def Motor; }`, `package System { private import Components::*; part secondaryBattery : Ba; }`, []string{"Motor"}, []string{"Battery"}},
		{"public-reexport", `package Components { part def Battery; } package Library { public import Components::*; }`, `package System { private import Library::*; part secondaryBattery : Ba; }`, []string{"Battery"}, nil},
		{"private-reexport", `package Components { part def Battery; } package Library { private import Components::*; }`, `package System { private import Library::*; part secondaryBattery : Ba; }`, nil, []string{"Battery"}},
		{"clash", `package Left { part def Battery; } package Right { part def Battery; }`, `package System { private import Left::*; private import Right::*; part secondaryBattery : Ba; }`, nil, []string{"Battery"}},
		{"clash-reversed", `package Left { part def Battery; } package Right { part def Battery; }`, `package System { private import Right::*; private import Left::*; part secondaryBattery : Ba; }`, nil, []string{"Battery"}},
		{"same-element", `package Components { part def Battery; } package Library { public import Components::*; }`, `package System { private import Components::*; private import Library::*; part secondaryBattery : Ba; }`, []string{"Battery"}, nil},
		{"library-import", `package Components;`, `package System { private import ScalarValues::*; attribute value : Re; part secondaryBattery : Ba; }`, []string{"Real"}, nil},
		{"inherited-import", `package Components { part def Battery; } package Types { part def Base { protected import Components::*; } }`, `package System { part def Derived :> Types::Base { part secondaryBattery : Ba; } }`, []string{"Battery"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ws := model.NewWorkspace()
			server := NewServer(ws)
			root := t.TempDir()
			// Definitions are on-disk workspace content, never an open editor buffer.
			ws.SetOnDisk(filepath.Join(root, "definitions.sysml"), []byte(tc.definitions))
			name := filepath.Join(root, "system.sysml")
			ws.Open(name, []byte(tc.consumer), 1)
			offset := strings.Index(tc.consumer, "secondaryBattery : Ba") + len("secondaryBattery : Ba")
			list, err := server.Completion(context.Background(), &protocol.CompletionParams{
				TextDocumentPositionParams: protocol.TextDocumentPositionParams{
					TextDocument: protocol.TextDocumentIdentifier{URI: uri.File(name)},
					Position:     offsetToPosition([]byte(tc.consumer), offset),
				},
			})
			if err != nil || list == nil {
				t.Fatalf("completion failed: %v", err)
			}
			items := map[string]protocol.CompletionItem{}
			counts := map[string]int{}
			for _, item := range list.Items {
				items[item.Label] = item
				counts[item.Label]++
			}
			for _, label := range tc.want {
				item, ok := items[label]
				if !ok {
					t.Errorf("missing imported completion %q", label)
					continue
				}
				if counts[label] != 1 {
					t.Errorf("completion %q offered %d times", label, counts[label])
				}
				if item.Kind == protocol.CompletionItemKindKeyword {
					t.Errorf("declaration %q misclassified as keyword", label)
				}
			}
			if strings.Contains(tc.name, "clash-control") && items["Battery"].Detail != "item def" {
				t.Errorf("completion chose an imported binding instead of the native owned/outer binding: %v", items["Battery"])
			}
			for _, label := range tc.absent {
				if _, ok := items[label]; ok {
					t.Errorf("hidden name %q offered", label)
				}
			}
		})
	}
}

func TestCompletionRefreshesAfterImportEdits(t *testing.T) {
	ws := model.NewWorkspace()
	server := NewServer(ws)
	root := t.TempDir()
	ws.SetOnDisk(filepath.Join(root, "definitions.sysml"), []byte(`package Components { part def Battery; }`))
	name := filepath.Join(root, "system.sysml")
	for version, imported := range []bool{true, false, true} {
		imports := ""
		if imported {
			imports = "private import Components::*;"
		}
		content := "package System { " + imports + " part secondaryBattery : Ba; }"
		ws.Open(name, []byte(content), version+1)
		list, err := server.Completion(context.Background(), &protocol.CompletionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{URI: uri.File(name)},
				Position:     offsetToPosition([]byte(content), strings.Index(content, "secondaryBattery : Ba")+len("secondaryBattery : Ba")),
			},
		})
		if err != nil || list == nil {
			t.Fatalf("completion failed: %v", err)
		}
		found := false
		for _, item := range list.Items {
			if item.Label == "Battery" {
				found = true
			}
		}
		if found != imported {
			t.Errorf("import present=%v but Battery completion=%v", imported, found)
		}
	}
}

func TestCompletionThenTypingKeepsCrossFileNavigation(t *testing.T) {
	ws := model.NewWorkspace()
	server := NewServer(ws)
	root := t.TempDir()
	ws.SetOnDisk(filepath.Join(root, "definitions.sysml"), []byte(`package Components { port def PowerPort; part def Battery { port output : PowerPort; } }`))
	name := filepath.Join(root, "system.sysml")
	original := "package System {\n    private import Components::*;\n    part battery : Battery;\n}\n"
	ws.Open(name, []byte(original), 1)
	partial := strings.Replace(original, "}\n", "    part secondaryBattery : Ba;\n}\n", 1)
	ws.Update(name, []byte(partial), 2)
	offset := strings.Index(partial, "secondaryBattery : Ba") + len("secondaryBattery : Ba")
	params := protocol.TextDocumentPositionParams{TextDocument: protocol.TextDocumentIdentifier{URI: uri.File(name)}, Position: offsetToPosition([]byte(partial), offset)}
	if _, err := server.Completion(context.Background(), &protocol.CompletionParams{TextDocumentPositionParams: params}); err != nil {
		t.Fatal(err)
	}
	accepted := strings.Replace(partial, "secondaryBattery : Ba;", "secondaryBattery : Battery;", 1)
	ws.Update(name, []byte(accepted), 3)
	locations, err := server.Definition(context.Background(), &protocol.DefinitionParams{TextDocumentPositionParams: params})
	if err != nil || len(locations) != 1 || !strings.HasSuffix(locations[0].URI.Filename(), "definitions.sysml") {
		t.Fatalf("accepted completion has no cross-file definition: %v %v", locations, err)
	}
}
