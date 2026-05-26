package dsl

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
		if !isDigit(data[start]) {
			s.Out <- illegalToken(cursor, start+1, string(data[start:start+1]))
			continue
		}

		end := start + 1
		for end < len(data) && isDigit(data[end]) {
			end++
		}

		s.Out <- newToken(TokInt, cursor, end, string(data[start:end]))
	}
}
