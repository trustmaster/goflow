package dsl

import (
	"bytes"
	"fmt"
)

// TokenType differentiates tokens
type TokenType string

const (
	// Token types
	tokIllegal    = TokenType("illegal")    // S:Fallback P:9
	tokBeginFile  = TokenType("beginFile")  // S:Auto
	tokEOF        = TokenType("eof")        // S:Auto
	tokWhitespace = TokenType("whitespace") // S:Chars P:1
	tokEOL        = TokenType("eol")        // S:Chars P:1
	tokComment    = TokenType("comment")    // # S:Comment P:2

	// Literals
	tokIdent     = TokenType("ident")     // ProcName S:Chars P:3
	tokInt       = TokenType("int")       // 123 S:Chars P:2
	tokQuotedStr = TokenType("quotedStr") // "data" S:Quoted P:2

	// Operators
	tokEqual  = TokenType("equal")  // = S:Keyword P:2
	tokDot    = TokenType("dot")    // . S:Keyword P:2
	tokColon  = TokenType("colon")  // : S:Keyword P:2
	tokLparen = TokenType("lparen") // ( S:Keyword P:2
	tokRparen = TokenType("rparen") // ) S:Keyword P:2
	tokArrow  = TokenType("arrow")  // -> S:Keyword P:2
	tokSlash  = TokenType("slash")  // / S:Keyword P:2

	// Keywords
	tokInport  = TokenType("inport")  // S:Keyword P:2
	tokOutport = TokenType("outport") // S:Keyword P:2
)

// Token represents a single lexem in a File
type Token struct {
	Type  TokenType
	File  *File
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
	File  <-chan *File
	Token chan<- Token
	Err   chan<- LexError
}

// Process scans the input stream and splits it into tokens
func (c *Tokenizer) Process() {
	for f := range c.File {
		// Send the new file
		pos := 0

		c.Token <- Token{
			Type:  tokBeginFile,
			File:  f,
			Pos:   pos,
			Value: f.Name,
		}

		// var t Token
		// l := 0
		// fileSize := len(f.Data)

		// for ; pos < fileSize; pos++ {
		// 	ch := f.Data[pos]
		// 	if isWhitespace(ch) {
		// 		t, l = scanByClass(r, ch, WS, pos, isWhitespace)
		// 	} else if isLineBreak(ch) {
		// 		t, l = scanByClass(r, ch, EOL, pos, isLineBreak)
		// 	} else if isLetter(ch) {
		// 		// Can be a start of an ident or keyword
		// 		t, l = scanByClass(r, ch, IDENT, pos, isIdent)
		// 		// Check for keywords
		// 		val := strings.ToLower(t.Value)
		// 		switch val {
		// 		case "inport":
		// 			t.Type = INPORT
		// 		case "outport":
		// 			t.Type = OUTPORT
		// 		}
		// 	} else if isDigit(ch) {
		// 		t, l = scanByClass(r, ch, INT, pos, isDigit)
		// 	}
		// 	c.Token <- t
		// 	pos += l + 1
		// }

		c.Token <- Token{
			Type:  tokEOF,
			Pos:   pos,
			Value: f.Name,
		}
	}
}

// predicate checks if a char belongs to a class
type predicate func(ch byte) bool

// scanByClass scans all characters belonging to the same class into a token
func scanByClass(data []byte, first byte, tt TokenType, pos int, belongs predicate) (Token, int) {
	buf := bytes.NewBufferString(string(first))
	len := 0
	// for ; pos{
	// 	ch := data[pos]
	// 	if ch == rune(0) || err != nil {
	// 		break
	// 	} else if belongs(ch) {
	// 		buf.WriteRune(ch)
	// 		len++
	// 	} else {
	// 		r.UnreadRune()
	// 		break
	// 	}
	// }
	return Token{
		Type:  tt,
		Pos:   pos,
		Value: buf.String(),
	}, len
}

func isWhitespace(ch rune) bool {
	return ch == '\t' || ch == ' '
}

func isLineBreak(ch rune) bool {
	return ch == '\r' || ch == '\n'
}

func isLetter(ch rune) bool {
	return ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isIdent(ch rune) bool {
	return isLetter(ch) || isDigit(ch) || ch == '_'
}
