package dsl

import (
	"bufio"
	"bytes"
	"fmt"
)

// TokenType differentiates tokens
type TokenType int

const (
	// Token types
	ILLEGAL TokenType = iota
	BOF               // Beginning of file
	EOF               // End of file
	WS                // Whitespace
	EOL               // End of line

	// Literals
	LITERAL_BEG
	IDENT  // ProcName
	INT    // 123
	FLOAT  // 123.45
	STRING // "data"
	LITERAL_END

	// Operators
	OPS_BEG
	EQL    // =
	DOT    // .
	COL    // :
	LPAREN // (
	RPAREN // )
	ARROW  // ->
	HASH   // #
	OPS_END

	// Keywords
	KEYWORD_BEG
	INPORT
	OUTPORT
	KEYWORD_END
)

// Token represents a single lexem
type Token struct {
	Type  TokenType
	Pos   int
	Value string
}

func (t Token) String() string {
	return t.Value
}

// LexError is a lexical error
type LexError struct {
	File string
	Pos  int
	Err  error
}

// Error returns an error message
func (e LexError) Error() string {
	return fmt.Sprintf("Error scanning file '%s' at pos %d: %s", e.File, e.Pos, e.Err.Error())
}

// Tokenizer splits the file into tokens
type Tokenizer struct {
	File  <-chan File
	Token chan<- Token
	Error chan<- LexError
}

// Process scans the input stream and splits it into tokens
func (c *Tokenizer) Process() {
	for f := range c.File {
		// Send the new file
		pos := 0

		c.Token <- Token{
			Type:  BOF,
			Pos:   pos,
			Value: f.Name,
		}

		r := bufio.NewReader(f.Reader)
		var t Token
		len := 0

		for {
			ch, _, err := r.ReadRune()
			if ch == rune(0) || err != nil {
				break
			}
			if isWhitespace(ch) {
				t = Token{
					Type:  WS,
					Pos:   pos,
					Value: string(ch),
				}
				len = scanClass(r, &t, isWhitespace)
			} else if isLineBreak(ch) {
				t = Token{
					Type:  EOL,
					Pos:   pos,
					Value: string(ch),
				}
				len = scanClass(r, &t, isLineBreak)
			}
			c.Token <- t
			pos += len + 1
		}

		c.Token <- Token{
			Type:  EOF,
			Pos:   pos,
			Value: f.Name,
		}
	}
}

// predicate checks fi a char belongs to a class
type predicate func(ch rune) bool

// scanClass scans all characters belonging to the same class
func scanClass(r *bufio.Reader, t *Token, belongs predicate) int {
	buf := bytes.NewBufferString(t.Value)
	len := 0
	for {
		ch, _, err := r.ReadRune()
		if ch == rune(0) || err != nil {
			break
		} else if belongs(ch) {
			buf.WriteRune(ch)
			len++
		} else {
			r.UnreadRune()
			break
		}
	}
	t.Value = buf.String()
	return len
}

func isWhitespace(ch rune) bool {
	return ch == '\t' || ch == ' '
}

func isLineBreak(ch rune) bool {
	return ch == '\r' || ch == '\n'
}
