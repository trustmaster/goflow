package dsl

// StripTrivia removes whitespace and comment tokens from a token stream.
type StripTrivia struct {
	In  <-chan Token
	Out chan<- Token
}

// Process forwards all non-trivia tokens.
func (s *StripTrivia) Process() {
	for tok := range s.In {
		if tok.Type == TokWhitespace || tok.Type == TokComment {
			continue
		}

		s.Out <- tok
	}
}
