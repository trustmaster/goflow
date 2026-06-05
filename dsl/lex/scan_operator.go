package lex

import "github.com/trustmaster/goflow/dsl/types"

// ScanOperatorToken scans punctuation and multi-character operator tokens.
type ScanOperatorToken struct {
	In  <-chan types.Cursor
	Out chan<- types.Token
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
		kind := types.TokIllegal

		switch data[start] {
		case '=':
			kind = types.TokEqual
		case '.':
			kind = types.TokDot
		case ':':
			kind = types.TokColon
		case '(':
			kind = types.TokLParen
		case ')':
			kind = types.TokRParen
		case '[':
			kind = types.TokLBracket
		case ']':
			kind = types.TokRBracket
		case '/':
			kind = types.TokSlash
		case '-':
			if end < len(data) && data[end] == '>' {
				kind = types.TokArrow
				end++
			}
		}

		value := string(data[start:end])
		if kind == types.TokIllegal {
			s.Out <- types.IllegalToken(cursor, end, value)
			continue
		}

		s.Out <- types.NewToken(kind, cursor, end, value)
	}
}
