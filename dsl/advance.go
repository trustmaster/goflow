package dsl

// Advance emits scanned tokens and advances the lexer cursor.
type Advance struct {
	In   <-chan Token
	Next chan<- Cursor
	Out  chan<- Token
}

// Process forwards tokens and emits the next cursor for non-EOF tokens.
func (a *Advance) Process() {
	for tok := range a.In {
		a.Out <- tok

		if tok.Type == TokEOF {
			return
		}

		cursor := Cursor{
			File:   tok.File,
			Offset: tok.Pos,
			Line:   tok.Span.Line,
			Column: tok.Span.Column,
		}
		a.Next <- advanceCursor(cursor, tok.Span.End)
	}
}
