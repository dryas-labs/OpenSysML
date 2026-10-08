// SPDX-License-Identifier: Apache-2.0
// Added by DRYAS maintainers for the downstream OpenSysML implementation.
// DRYAS read-only bridge to existing native calculations. No semantic rule changes.
package semantics

import "github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"

func (m *Model) DryasSemanticMetadataBases(sym *symbols.Symbol) ([]*symbols.Symbol, bool) {
	return m.semanticMetadataBases(sym)
}

func (m *Model) DryasIsSemanticMetadata(sym *symbols.Symbol) bool {
	return m.isSemanticMetadata(sym)
}
