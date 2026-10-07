package model

import (
	"sort"
	"strings"

	"github.com/Open-MBEE/OpenSysML/internal/semantic/resolve"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/semantics"
	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
)

// CompletionCandidate retains the spelling that is visible at the cursor,
// which may be an alias or short name rather than the declaration's name.
type CompletionCandidate struct {
	Name string
	Sym  *symbols.Symbol
}

// CompletionCandidates offers unqualified names using the same lookup as a
// written reference. Enumeration supplies spellings; lookup decides which
// binding wins and rejects hidden imported names. No import is resolved by
// completion-specific precedence rules.
func (w *Workspace) CompletionCandidates(scope *symbols.Scope) []CompletionCandidate {
	if scope == nil {
		return nil
	}
	doc := symbols.DocNameOf(scope)
	top := w.TopLevelSymbols(doc)
	names := map[string]bool{}
	var roots []string
	for _, sym := range top {
		if sym == nil {
			continue
		}
		roots = append(roots, strings.SplitN(sym.Name, "::", 2)[0])
		names[leafName(sym.Name)] = true
		if sym.ShortName != "" {
			names[sym.ShortName] = true
		}
	}
	for _, visible := range w.VisibleNames(scope, VisibleNamesOptions{MaxDepth: 1, LibraryRoots: roots}) {
		names[visible.Name] = true
	}
	ordered := make([]string, 0, len(names))
	for name := range names {
		ordered = append(ordered, name)
	}
	sort.Strings(ordered)
	w.mu.Lock()
	defer w.mu.Unlock()
	var out []CompletionCandidate
	w.queryLocked(doc, func(r *resolve.Resolver, _ *semantics.Model) {
		for _, name := range ordered {
			if sym, ok := r.LookupName(scope, name); ok && sym != nil {
				out = append(out, CompletionCandidate{Name: name, Sym: sym})
			}
		}
	})
	return out
}
