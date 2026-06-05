package parse

import "github.com/trustmaster/goflow/dsl/types"

// SegmentStatements groups a token stream into newline-delimited statements.
type SegmentStatements struct {
	In  <-chan types.Token
	Out chan<- types.Statement
}

// Process emits one Statement per non-empty line and flushes the final statement at EOF.
func (s *SegmentStatements) Process() {
	var tokens []types.Token

	flush := func() {
		if len(tokens) == 0 {
			return
		}

		s.Out <- types.NewStatement(tokens)

		tokens = nil
	}

	for tok := range s.In {
		//nolint:exhaustive // default handles all non-delimiter tokens
		switch tok.Type {
		case types.TokEOL:
			flush()
		case types.TokEOF:
			flush()

			return
		default:
			tokens = append(tokens, tok)
		}
	}

	flush()
}
