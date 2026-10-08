// SPDX-License-Identifier: Apache-2.0
// Added by DRYAS maintainers for the downstream OpenSysML implementation.
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"github.com/Open-MBEE/OpenSysML/internal/semantic/resolve"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/semantics"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
)

// ParsedSnapshot exposes this snapshot's native syntax for a selected editor
// presentation. It does not replace record-backed index symbols or invalidate
// the workspace. Completion enumeration need not parse every candidate's file.
func (d *Document) ParsedSnapshot() *Document {
	if d == nil || !d.Recorded() {
		return d
	}
	return newDocument(d.Name, d.Content, d.Version)
}

// CompletionTarget follows an alias through the same resolver used by model
// references. Ordinary symbols and unresolved aliases retain their declaration.
func (w *Workspace) CompletionTarget(sym *symbols.Symbol) *symbols.Symbol {
	if sym == nil || sym.Kind != symbols.SymbolAlias {
		return sym
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	out := sym
	w.queryLocked(sym.DocName, func(r *resolve.Resolver, _ *semantics.Model) {
		if target, ok := r.ResolveAliasTarget(sym); ok && target != nil {
			out = target
		}
	})
	return out
}

// CompletionSnapshot retains immutable workspace documents for one offered list.
// It carries no new semantic index or name-resolution rules.
type CompletionSnapshot struct {
	generation uint64
	documents  map[string]*Document
}

// CompletionSnapshot captures documents and generation under the same read lock.
func (w *Workspace) CompletionSnapshot() *CompletionSnapshot {
	w.mu.RLock()
	defer w.mu.RUnlock()
	snapshot := &CompletionSnapshot{generation: w.generation, documents: make(map[string]*Document, len(w.docs))}
	for name, doc := range w.docs {
		snapshot.documents[name] = doc
	}
	return snapshot
}

// Generation is the generation captured before enumerating candidates.
func (snapshot *CompletionSnapshot) Generation() uint64 { return snapshot.generation }

// CompletionUnchangedExcept reports whether only the named document may have
// changed. The caller must check that document's exact permitted text edit.
// A changed analysis configuration with no source edit also invalidates the list.
func (w *Workspace) CompletionUnchangedExcept(snapshot *CompletionSnapshot, name string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if snapshot.generation == w.generation {
		return true
	}
	if len(snapshot.documents) != len(w.docs) || snapshot.documents[name] == w.docs[name] {
		return false
	}
	for file, doc := range snapshot.documents {
		if file != name && w.docs[file] != doc {
			return false
		}
	}
	return w.docs[name] != nil
}
