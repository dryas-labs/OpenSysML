// SPDX-License-Identifier: Apache-2.0
package lsp

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"path/filepath"
	"strings"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"

	"github.com/Open-MBEE/OpenSysML/internal/semantic/symbols"
	"github.com/Open-MBEE/OpenSysML/internal/workspace/model"
)

const completionBatchLimit = 8

type completionResolveData struct {
	Token string `json:"token"`
	Index int    `json:"index"`
}

type completionSource struct {
	symbol   *symbols.Symbol
	document *model.Document
}

type completionEntry struct {
	label, reference string
	original, target completionSource
}

type completionBatch struct {
	context completionContext
	token   string
	entries map[int]completionEntry
}

func (s *Server) lazyCompletion() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.completionResolve
}

func clientResolvesCompletionDocumentation(caps protocol.ClientCapabilities) bool {
	if caps.TextDocument == nil || caps.TextDocument.Completion == nil || caps.TextDocument.Completion.CompletionItem == nil {
		return false
	}
	item := caps.TextDocument.Completion.CompletionItem
	if item.ResolveSupport == nil {
		return false
	}
	for _, property := range item.ResolveSupport.Properties {
		if property == "documentation" {
			return true
		}
	}
	return false
}

func (s *Server) finishCompletion(ctx context.Context, c *completionItems) (*protocol.CompletionList, error) {
	if ctx.Err() != nil {
		return nil, protocol.ErrRequestCancelled
	}
	result := c.list()
	if !c.lazy || len(c.entries) == 0 {
		return result, nil
	}
	if s.ws.Generation() != c.state.Generation() {
		return nil, staleCompletion()
	}
	batch := completionBatch{context: c.context, token: rand.Text(), entries: map[int]completionEntry{}}
	for index := range result.Items {
		item := &result.Items[index]
		if entry, ok := c.entries[item.Label]; ok {
			batch.entries[index] = entry
			item.Data = completionResolveData{Token: batch.token, Index: index}
		}
	}
	s.mu.Lock()
	s.completionBatches = append(s.completionBatches, batch)
	if len(s.completionBatches) > completionBatchLimit {
		s.completionBatches[0] = completionBatch{}
		s.completionBatches = s.completionBatches[1:]
	}
	s.mu.Unlock()
	return result, nil
}

func staleCompletion() error {
	return jsonrpc2.Errorf(protocol.CodeContentModified, "The completion is no longer current; request completions again.")
}

// CompletionResolve adds presentation only: insertion, filtering and sorting
// fields remain exactly as the client received them. Handles bind to the native
// declarations originally offered, never to a fresh lookup of the display name.
func (s *Server) CompletionResolve(ctx context.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
	if ctx.Err() != nil {
		return nil, protocol.ErrRequestCancelled
	}
	if params == nil {
		return nil, jsonrpc2.Errorf(jsonrpc2.InvalidParams, "Missing completion item.")
	}
	result := *params
	if params.Data == nil {
		return &result, nil
	}
	var data completionResolveData
	raw, err := json.Marshal(params.Data)
	if err != nil || json.Unmarshal(raw, &data) != nil || data.Token == "" || data.Index < 0 {
		return nil, jsonrpc2.Errorf(jsonrpc2.InvalidParams, "Invalid completion handle.")
	}
	var state completionContext
	var entry completionEntry
	var found bool
	s.mu.Lock()
	for _, batch := range s.completionBatches {
		if batch.token == data.Token {
			entry, found = batch.entries[data.Index]
			state = batch.context
			break
		}
	}
	s.mu.Unlock()
	if !found || !s.completionContextCurrent(state) {
		return nil, staleCompletion()
	}
	if params.Label != entry.label {
		return nil, jsonrpc2.Errorf(jsonrpc2.InvalidParams, "Completion label does not match its handle.")
	}
	original, originalDoc, err := s.completionDeclaration(entry.original, state)
	if err != nil {
		return nil, err
	}
	if original.Kind == symbols.SymbolAlias {
		target := s.ws.CompletionTarget(original)
		if target == nil || target.DocName != entry.target.symbol.DocName || target.DeclSpan != entry.target.symbol.DeclSpan {
			return nil, staleCompletion()
		}
	}
	sym, doc := original, originalDoc
	if entry.original.symbol != entry.target.symbol {
		sym, doc, err = s.completionDeclaration(entry.target, state)
		if err != nil {
			return nil, err
		}
	}
	if ctx.Err() != nil {
		return nil, protocol.ErrRequestCancelled
	}
	signature := completionSignature(doc, sym)
	declaration := symbols.FQNOf(sym)
	file := filepath.Base(sym.DocName)
	comments := declarationDocComments(doc, sym)
	if s.wantsMarkdownCompletion() {
		value := markdownBlock(signature) + "\n\nCompletion: " + markdownInline(entry.reference) +
			"  \nDeclaration: " + markdownInline(declaration) + "  \nSource: " + markdownInline(file)
		if prose := docCommentProse(comments); prose != "" {
			value += "\n\n---\n\n" + prose
		}
		result.Documentation = protocol.MarkupContent{Kind: protocol.Markdown, Value: value}
	} else {
		value := signature + "\n\nCompletion: " + entry.reference + "\nDeclaration: " + declaration + "\nSource: " + file
		if prose := docCommentProse(comments); prose != "" {
			value += "\n\n" + prose
		}
		result.Documentation = protocol.MarkupContent{Kind: protocol.PlainText, Value: value}
	}
	if ctx.Err() != nil {
		return nil, protocol.ErrRequestCancelled
	}
	if !s.completionContextCurrent(state) || !s.completionSourceCurrent(entry.original, state) || !s.completionSourceCurrent(entry.target, state) {
		return nil, staleCompletion()
	}
	return &result, nil
}

// A changed declaration or enclosing prefix invalidates its source identity.
// Typing after a declaration does not invalidate it, so a client can continue
// filtering a cached list while the user extends a reference's identifier.
func (s *Server) completionSourceCurrent(source completionSource, state completionContext) bool {
	if source.symbol == nil {
		return false
	}
	if source.document == nil {
		return s.ws.IsLibraryDocument(source.symbol.DocName)
	}
	current := s.ws.Document(source.symbol.DocName)
	if current == nil {
		return false
	}
	if current == source.document {
		return true
	}

	if state.document != nil && source.document.Name == state.document.Name {
		span := source.symbol.DeclSpan
		return span.End() <= state.wordStart || span.Offset >= state.offset
	}
	end := source.symbol.DeclSpan.End()
	return end > 0 && end <= len(current.Content) && end <= len(source.document.Content) &&
		bytes.Equal(current.Content[:end], source.document.Content[:end])
}

func (s *Server) completionDeclaration(source completionSource, state completionContext) (*symbols.Symbol, *model.Document, error) {
	if !s.completionSourceCurrent(source, state) {
		return nil, nil, staleCompletion()
	}
	doc := source.document
	if doc == nil {
		doc = s.document(source.symbol.DocName)
	}
	if doc == nil {
		return nil, nil, staleCompletion()
	}
	doc = doc.ParsedSnapshot()
	parsed := symbolAtOffset(doc.Scope, source.symbol.DeclSpan.Offset)
	if parsed == nil || parsed.DeclSpan != source.symbol.DeclSpan {
		return nil, nil, staleCompletion()
	}
	return parsed, doc, nil
}

func completionSignature(doc *model.Document, sym *symbols.Symbol) string {
	span := sym.DeclSpan
	end := span.End()
	if body, ok := bodyOf(doc.Content, span); ok {
		end = body.Offset
	}
	if span.Offset < 0 || end > len(doc.Content) || span.Offset >= end {
		return sym.Notation() + " " + sym.Name
	}
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(string(doc.Content[span.Offset:end])), ";"))
}

func markdownFence(text string, minimum int) string {
	longest, run := 0, 0
	for _, char := range text {
		if char == '`' {
			run++
			if run > longest {
				longest = run
			}
		} else {
			run = 0
		}
	}
	if longest+1 > minimum {
		minimum = longest + 1
	}
	return strings.Repeat("`", minimum)
}

func markdownInline(text string) string {
	fence := markdownFence(text, 1)
	if strings.HasPrefix(text, "`") || strings.HasSuffix(text, "`") {
		text = " " + text + " "
	}
	return fence + text + fence
}

func markdownBlock(text string) string {
	fence := markdownFence(text, 3)
	return fence + "sysml\n" + text + "\n" + fence
}

type completionContext struct {
	document          *model.Document
	state             *model.CompletionSnapshot
	wordStart, offset int
}

// A client filters a complete list as an identifier is typed. Permit exactly
// that edit, while rejecting import changes, other model edits and new files.
func (s *Server) completionContextCurrent(state completionContext) bool {
	if state.document == nil || state.state == nil {
		return false
	}
	if !s.ws.CompletionUnchangedExcept(state.state, state.document.Name) {
		return false
	}
	current := s.document(state.document.Name)
	if current == nil {
		return false
	}
	if current == state.document {
		return true
	}
	before, after := state.document.Content[:state.wordStart], state.document.Content[state.offset:]
	content := current.Content
	if len(content) < len(before)+len(after) || !bytes.HasPrefix(content, before) || !bytes.HasSuffix(content, after) {
		return false
	}
	for _, b := range content[len(before) : len(content)-len(after)] {
		if !isIdentByte(b) {
			return false
		}
	}
	return true
}
