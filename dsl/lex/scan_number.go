package lex

import (
	"unicode/utf8"

	"github.com/trustmaster/goflow/dsl/types"
)

// ScanNumberToken scans an integer token.
type ScanNumberToken struct {
	In  <-chan types.Cursor
	Out chan<- types.Token
}

// Process scans integers from each incoming cursor.
func (s *ScanNumberToken) Process() {
	for cursor := range s.In {
		if cursor.File == nil || cursor.Offset >= len(cursor.File.Data) {
			continue
		}

		data := cursor.File.Data

		start := cursor.Offset

		r, size := utf8.DecodeRune(data[start:])
		if !types.IsDigit(r) {
			s.Out <- types.IllegalToken(cursor, start+size, string(data[start:start+size]))
			continue
		}

		end := start + size
		for end < len(data) {
			r, size := utf8.DecodeRune(data[end:])
			if !types.IsDigit(r) {
				break
			}

			end += size
		}

		s.Out <- types.NewToken(types.TokInt, cursor, end, string(data[start:end]))
	}
}
