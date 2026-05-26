package dsl

// ScanOperatorToken scans punctuation and multi-character operator tokens.
type ScanOperatorToken struct {
	In  <-chan Cursor
	Out chan<- Token
}

// Process scans operators from each incoming cursor.
func (s *ScanOperatorToken) Process() {
	for cursor := range s.In {
		if cursor.File == nil || cursor.Offset >= len(cursor.File.Data) {
			continue
		}

		data := cursor.File.Data
		start := cursor.Offset
		end := start + 1
		kind := TokIllegal

		switch data[start] {
		case '=':
			kind = TokEqual
		case '.':
			kind = TokDot
		case ':':
			kind = TokColon
		case '(':
			kind = TokLParen
		case ')':
			kind = TokRParen
		case '[':
			kind = TokLBracket
		case ']':
			kind = TokRBracket
		case '/':
			kind = TokSlash
		case '-':
			if end < len(data) && data[end] == '>' {
				kind = TokArrow
				end++
			}
		}

		value := string(data[start:end])
		if kind == TokIllegal {
			s.Out <- illegalToken(cursor, end, value)
			continue
		}

		s.Out <- newToken(kind, cursor, end, value)
	}
}
