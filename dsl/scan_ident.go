package dsl

import "unicode/utf8"

// ScanIdentToken scans an identifier token.
type ScanIdentToken struct {
	In  <-chan Cursor
	Out chan<- Token
}

// Process scans identifiers from each incoming cursor.
func (s *ScanIdentToken) Process() {
	for cursor := range s.In {
		if cursor.File == nil || cursor.Offset >= len(cursor.File.Data) {
			continue
		}

		data := cursor.File.Data
		start := cursor.Offset

		r, size := utf8.DecodeRune(data[start:])
		if !isIdentStart(r) {
			s.Out <- illegalToken(cursor, start+size, string(data[start:start+size]))
			continue
		}

		end := start + size
		for end < len(data) {
			r, size := utf8.DecodeRune(data[end:])
			if !isIdentPart(r) {
				break
			}
			end += size
		}

		s.Out <- newToken(TokIdent, cursor, end, string(data[start:end]))
	}
}
