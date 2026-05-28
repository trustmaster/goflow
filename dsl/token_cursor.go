package dsl

import "fmt"

// tokenCursor is a position-tracking helper used by statement parsers.
type tokenCursor struct {
	tokens []Token
	pos    int
}

func newTokenCursor(tokens []Token) *tokenCursor {
	return &tokenCursor{tokens: tokens}
}

// peek returns the current token without advancing.
func (c *tokenCursor) peek() Token {
	if c.pos >= len(c.tokens) {
		return Token{Type: TokEOF}
	}

	return c.tokens[c.pos]
}

// consume returns the current token and advances one position.
func (c *tokenCursor) consume() Token {
	tok := c.peek()
	if c.pos < len(c.tokens) {
		c.pos++
	}

	return tok
}

// expect consumes and returns the current token if it matches t, otherwise returns a ParseError.
func (c *tokenCursor) expect(t TokenType) (Token, *ParseError) {
	tok := c.peek()
	if tok.Type != t {
		return tok, &ParseError{
			Span: tok.Span,
			Err:  fmt.Errorf("expected %s, got %s %q", t, tok.Type, tok.Value),
		}
	}

	c.pos++

	return tok, nil
}

// expectIdent consumes and returns the current token if it is an identifier.
func (c *tokenCursor) expectIdent() (Token, *ParseError) {
	tok := c.peek()
	if tok.Type != TokIdent {
		return tok, &ParseError{
			Span: tok.Span,
			Err:  fmt.Errorf("expected identifier, got %s %q", tok.Type, tok.Value),
		}
	}

	c.pos++

	return tok, nil
}

// done returns true when all tokens have been consumed.
func (c *tokenCursor) done() bool {
	return c.pos >= len(c.tokens)
}
