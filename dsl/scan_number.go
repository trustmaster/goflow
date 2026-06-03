package dsl

import "unicode/utf8"

// ScanNumberToken scans an integer token.
type ScanNumberToken struct {
	In  <-chan Cursor
	Out chan<- Token
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
		if !isDigit(r) {
			s.Out <- illegalToken(cursor, start+size, string(data[start:start+size]))
			continue
		}

		end := start + size
		for end < len(data) {
			r, size := utf8.DecodeRune(data[end:])
			if !isDigit(r) {
				break
			}
			end += size
		}

		s.Out <- newToken(TokInt, cursor, end, string(data[start:end]))
	}
}
