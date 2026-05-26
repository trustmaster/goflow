package dsl

import "bytes"

// ScanQuotedToken scans a single- or double-quoted string token.
type ScanQuotedToken struct {
	In  <-chan Cursor
	Out chan<- Token
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

func scanQuotedToken(cursor Cursor) Token {
	data := cursor.File.Data
	start := cursor.Offset

	quote := rune(data[start])
	if quote != '\'' && quote != '"' {
		return illegalToken(cursor, start+1, string(data[start:start+1]))
	}

	escape := rune('\\')
	escaped := false
	buf := bytes.NewBufferString(string(quote))
	end := len(data)
	closed := false

	for i := start + 1; i < len(data); i++ {
		r := rune(data[i])
		end = i + 1

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

	return newToken(TokQuoted, cursor, end, buf.String())
}
