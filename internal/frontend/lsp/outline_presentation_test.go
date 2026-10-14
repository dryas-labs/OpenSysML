// SPDX-License-Identifier: Apache-2.0
// Added by DRYAS maintainers: native outline presentation regression.
package lsp

import (
	"context"
	"github.com/Open-MBEE/OpenSysML/internal/workspace/model"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"path/filepath"
	"testing"
)

func TestDocumentSymbolPresentation(t *testing.T) {
	ws := model.NewWorkspace()
	server := NewServer(ws)
	name := uri.File(filepath.Join(t.TempDir(), "model.sysml")).Filename()
	ws.Open(name, []byte(`package Example {
  part def Vehicle {
   attribute mass : ScalarValues::Real;
   part eng : Engine;
   ref part driver : Person;
   port p : ~Power;
   part both : Engine, Person;
  }
  attribute def VehicleStatus;
  part def Engine;
  part def Person;
  port def Power;
 }`), 1)
	read := func() map[string]string {
		result, err := server.DocumentSymbol(context.Background(), &protocol.DocumentSymbolParams{TextDocument: protocol.TextDocumentIdentifier{URI: uri.File(name)}})
		if err != nil {
			t.Fatal(err)
		}
		found := map[string]string{}
		var visit func(protocol.DocumentSymbol)
		visit = func(symbol protocol.DocumentSymbol) {
			found[symbol.Name] = symbol.Detail
			for _, child := range symbol.Children {
				visit(child)
			}
		}
		for _, symbol := range result {
			visit(symbol.(protocol.DocumentSymbol))
		}
		return found
	}
	found := read()
	for key, expected := range map[string]string{
		"Example": "package", "Vehicle": "part def", "mass": "attribute : ScalarValues::Real",
		"eng": "part : Engine", "driver": "ref part : Person", "VehicleStatus": "attribute def",
		"p": "port : ~Power", "both": "part : Engine, Person",
	} {
		if found[key] != expected {
			t.Errorf("%s detail = %q; want %q", key, found[key], expected)
		}
	}
	// A newer in-memory document must replace earlier presentation metadata.
	ws.Open(name, []byte(`package Example { attribute eng : ScalarValues::Integer; }`), 2)
	updated := read()
	if updated["eng"] != "attribute : ScalarValues::Integer" {
		t.Fatalf("stale outline: %v", updated)
	}
	if _, exists := updated["driver"]; exists {
		t.Fatal("removed declaration retained")
	}
}
