package lex

import (
	"unicode/utf8"

	"github.com/trustmaster/goflow/dsl/types"
)

// Dispatch routes a cursor to the correct lexer scanner based on the current byte.
type Dispatch struct {
	In         <-chan types.Cursor
	Whitespace chan<- types.Cursor
	Comment    chan<- types.Cursor
	Quoted     chan<- types.Cursor
	Ident      chan<- types.Cursor
	Number     chan<- types.Cursor
	Operator   chan<- types.Cursor
	Eof        chan<- types.Token
}

// Process dispatches each cursor to exactly one scanner family.
func (d *Dispatch) Process() {
	for cursor := range d.In {
		if cursor.File == nil {
			continue
		}

		if cursor.Offset >= len(cursor.File.Data) {
			d.Eof <- types.NewToken(types.TokEOF, cursor, cursor.Offset, cursor.File.Name)
			continue
		}

		ch, _ := utf8.DecodeRune(cursor.File.Data[cursor.Offset:])

		switch {
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n':
			d.Whitespace <- cursor
		case ch == '#':
			d.Comment <- cursor
		case ch == '\'' || ch == '"':
			d.Quoted <- cursor
		case types.IsIdentStart(ch):
			d.Ident <- cursor
		case types.IsDigit(ch):
			d.Number <- cursor
		default:
			d.Operator <- cursor
		}
	}
}
