package dsl

// ScanCommentToken scans a hash comment until the end of the line.
type ScanCommentToken struct {
	In  <-chan Cursor
	Out chan<- Token
}

// Process scans comments from each incoming cursor.
func (s *ScanCommentToken) Process() {
	for cursor := range s.In {
		if cursor.File == nil || cursor.Offset >= len(cursor.File.Data) {
			continue
		}

		data := cursor.File.Data
		start := cursor.Offset
		if data[start] != '#' {
			s.Out <- illegalToken(cursor, start+1, string(data[start:start+1]))
			continue
		}

		end := start
		for end < len(data) && data[end] != '\n' && data[end] != '\r' {
			end++
		}

		s.Out <- newToken(TokComment, cursor, end, string(data[start:end]))
	}
}
