package dsl

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
		if !isIdentStart(data[start]) {
			s.Out <- illegalToken(cursor, start+1, string(data[start:start+1]))
			continue
		}

		end := start + 1
		for end < len(data) && isIdentPart(data[end]) {
			end++
		}

		s.Out <- newToken(TokIdent, cursor, end, string(data[start:end]))
	}
}
