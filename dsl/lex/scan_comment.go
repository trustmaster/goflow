package lex

import "github.com/trustmaster/goflow/dsl/types"

// ScanCommentToken scans a hash comment until the end of the line.
type ScanCommentToken struct {
	In  <-chan types.Cursor
	Out chan<- types.Token
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
			s.Out <- types.IllegalToken(cursor, start+1, string(data[start:start+1]))
			continue
		}

		end := start
		for end < len(data) && data[end] != '\n' && data[end] != '\r' {
			end++
		}

		s.Out <- types.NewToken(types.TokComment, cursor, end, string(data[start:end]))
	}
}
