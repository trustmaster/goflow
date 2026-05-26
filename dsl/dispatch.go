package dsl

// Dispatch routes a cursor to the correct lexer scanner based on the current byte.
type Dispatch struct {
	In         <-chan Cursor
	Whitespace chan<- Cursor
	Comment    chan<- Cursor
	Quoted     chan<- Cursor
	Ident      chan<- Cursor
	Number     chan<- Cursor
	Operator   chan<- Cursor
	Eof        chan<- Token
}

// Process dispatches each cursor to exactly one scanner family.
func (d *Dispatch) Process() {
	for cursor := range d.In {
		if cursor.File == nil {
			continue
		}

		if cursor.Offset >= len(cursor.File.Data) {
			d.Eof <- newToken(TokEOF, cursor, cursor.Offset, cursor.File.Name)
			continue
		}

		switch ch := cursor.File.Data[cursor.Offset]; {
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n':
			d.Whitespace <- cursor
		case ch == '#':
			d.Comment <- cursor
		case ch == '\'' || ch == '"':
			d.Quoted <- cursor
		case isIdentStart(ch):
			d.Ident <- cursor
		case isDigit(ch):
			d.Number <- cursor
		default:
			d.Operator <- cursor
		}
	}
}
