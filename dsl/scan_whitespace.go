package dsl

// ScanWhitespaceToken scans spaces, tabs, and end-of-line tokens.
type ScanWhitespaceToken struct {
	In  <-chan Cursor
	Out chan<- Token
}

// Process scans whitespace from each incoming cursor.
func (s *ScanWhitespaceToken) Process() {
	for cursor := range s.In {
		if cursor.File == nil || cursor.Offset >= len(cursor.File.Data) {
			continue
		}

		data := cursor.File.Data
		start := cursor.Offset
		ch := data[start]

		switch ch {
		case '\r':
			end := start + 1
			if end < len(data) && data[end] == '\n' {
				end++
			}
			s.Out <- newToken(TokEOL, cursor, end, string(data[start:end]))
		case '\n':
			s.Out <- newToken(TokEOL, cursor, start+1, string(data[start:start+1]))
		default:
			end := start
			for end < len(data) && (data[end] == ' ' || data[end] == '\t') {
				end++
			}
			s.Out <- newToken(TokWhitespace, cursor, end, string(data[start:end]))
		}
	}
}
