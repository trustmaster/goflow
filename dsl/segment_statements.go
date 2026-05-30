package dsl

// SegmentStatements groups a token stream into newline-delimited statements.
type SegmentStatements struct {
	In  <-chan Token
	Out chan<- Statement
}

// Process emits one Statement per non-empty line and flushes the final statement at EOF.
func (s *SegmentStatements) Process() {
	var tokens []Token

	flush := func() {
		if len(tokens) == 0 {
			return
		}

		s.Out <- newStatement(tokens)

		tokens = nil
	}

	for tok := range s.In {
		//nolint:exhaustive // default handles all non-delimiter tokens
		switch tok.Type {
		case TokEOL, TokNewFile:
			flush()
		case TokEOF:
			flush()

			return
		default:
			tokens = append(tokens, tok)
		}
	}

	flush()
}
