package dsl

import (
	"fmt"
)

// TokenType differentiates tokens
type TokenType string

const (
	// Token types
	tokIllegal    = TokenType("illegal")    // S:Fallback P:9
	tokNewFile    = TokenType("newFile")    // S:Auto
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
			Type:  tokNewFile,
			File:  f,
			Pos:   pos,
			Value: f.Name,
		}

		// TODO repace this component with a graph

		c.Token <- Token{
			Type:  tokEOF,
			Pos:   pos,
			Value: f.Name,
		}
	}
}
