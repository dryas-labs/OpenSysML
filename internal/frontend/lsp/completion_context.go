package lsp

import (
	"github.com/Open-MBEE/OpenSysML/internal/syntax/lexer"
	"github.com/Open-MBEE/OpenSysML/internal/syntax/source"
)

// completionInComment follows the native token boundaries, including an
// unfinished block at EOF. Comment delimiters inside quoted names and strings
// are not comment tokens. The position immediately after a closing delimiter
// or a line-note terminator belongs to code again.
func completionInComment(content []byte, offset int) bool {
	scan := lexer.New(source.New("", content))
	for {
		token := scan.Next()
		if token.Kind == lexer.EOF || token.Span.Offset >= offset {
			return false
		}
		switch token.Kind {
		case lexer.RegularComment, lexer.MLNote, lexer.SLNote:
			end := token.Span.End()
			if offset < end {
				return true
			}
			if offset == end {
				if token.Unterminated {
					return true
				}
				if token.Kind == lexer.SLNote && end > 0 && content[end-1] != '\n' && content[end-1] != '\r' {
					return true
				}
			}
		}
		if token.Span.End() > offset {
			return false
		}
	}
}
