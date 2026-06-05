package lex

import "github.com/trustmaster/goflow/dsl/types"

// Advance emits scanned tokens and advances the lexer cursor.
type Advance struct {
	In   <-chan types.Token
	Next chan<- types.Cursor
	Out  chan<- types.Token
}

// Process forwards tokens and emits the next cursor for non-EOF tokens.
func (a *Advance) Process() {
	for tok := range a.In {
		a.Out <- tok

		if tok.Type == types.TokEOF {
			return
		}

		cursor := types.Cursor{
			File:   tok.File,
			Offset: tok.Pos,
			Line:   tok.Span.Line,
			Column: tok.Span.Column,
		}
		a.Next <- types.AdvanceCursor(cursor, tok.Span.End)
	}
}
