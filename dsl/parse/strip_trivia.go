package parse

import "github.com/trustmaster/goflow/dsl/types"

// StripTrivia removes whitespace and comment tokens from a token stream.
type StripTrivia struct {
	In  <-chan types.Token
	Out chan<- types.Token
}

// Process forwards all non-trivia tokens.
func (s *StripTrivia) Process() {
	for tok := range s.In {
		if tok.Type == types.TokWhitespace || tok.Type == types.TokComment {
			continue
		}

		s.Out <- tok
	}
}
