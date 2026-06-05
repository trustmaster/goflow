package lex

import (
	"bytes"
	"unicode/utf8"

	"github.com/trustmaster/goflow/dsl/types"
)

// ScanQuotedToken scans a single- or double-quoted string token.
type ScanQuotedToken struct {
	In  <-chan types.Cursor
	Out chan<- types.Token
}

// Process scans quoted string tokens from each incoming cursor.
func (s *ScanQuotedToken) Process() {
	for cursor := range s.In {
		if cursor.File == nil || cursor.Offset >= len(cursor.File.Data) {
			continue
		}

		s.Out <- scanQuotedToken(cursor)
	}
}

func scanQuotedToken(cursor types.Cursor) types.Token {
	data := cursor.File.Data
	start := cursor.Offset

	quote, size := utf8.DecodeRune(data[start:])
	if quote != '\'' && quote != '"' {
		return types.IllegalToken(cursor, start+size, string(data[start:start+size]))
	}

	escape := rune('\\')
	escaped := false
	buf := bytes.NewBufferString(string(quote))
	end := start + size
	closed := false

	for end < len(data) {
		r, rSize := utf8.DecodeRune(data[end:])
		end += rSize

		if r == escape {
			if escaped {
				buf.Truncate(buf.Len() - 1)

				escaped = false
			} else {
				buf.WriteRune(r)

				escaped = true

				continue
			}
		}

		if r == quote {
			if escaped {
				buf.Truncate(buf.Len() - 1)
			} else {
				buf.WriteRune(r)

				closed = true

				break
			}
		}

		buf.WriteRune(r)

		escaped = false
	}

	if !closed {
		end = len(data)
	}

	return types.NewToken(types.TokQuoted, cursor, end, buf.String())
}
