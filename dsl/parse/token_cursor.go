package parse

import (
	"fmt"

	"github.com/trustmaster/goflow/dsl/types"
)

// tokenCursor is a position-tracking helper used by statement parsers.
type tokenCursor struct {
	tokens []types.Token
	pos    int
}

func newTokenCursor(tokens []types.Token) *tokenCursor {
	return &tokenCursor{tokens: tokens}
}

// peek returns the current token without advancing.
func (c *tokenCursor) peek() types.Token {
	if c.pos >= len(c.tokens) {
		if len(c.tokens) > 0 {
			last := c.tokens[len(c.tokens)-1]
			// Calculate EOF position after the last token,
			// accounting for newlines in the final token value.
			line := last.Span.Line
			col := last.Span.Column

			for _, r := range last.Value {
				if r == '\n' {
					line++
					col = 1
				} else {
					col++
				}
			}

			return types.Token{
				Type: types.TokEOF,
				Span: types.Span{
					File:   last.Span.File,
					Offset: last.Span.End,
					Line:   line,
					Column: col,
					End:    last.Span.End,
				},
			}
		}

		return types.Token{Type: types.TokEOF}
	}

	return c.tokens[c.pos]
}

// consume returns the current token and advances one position.
func (c *tokenCursor) consume() types.Token {
	tok := c.peek()
	if c.pos < len(c.tokens) {
		c.pos++
	}

	return tok
}

// expect consumes and returns the current token if it matches t, otherwise returns a ParseError.
func (c *tokenCursor) expect(t types.TokenType) (types.Token, *types.ParseError) {
	tok := c.peek()
	if tok.Type != t {
		return tok, &types.ParseError{
			Span: tok.Span,
			Err:  fmt.Errorf("expected %s, got %s %q", t, tok.Type, tok.Value),
		}
	}

	c.pos++

	return tok, nil
}

// expectIdent consumes and returns the current token if it is an identifier.
func (c *tokenCursor) expectIdent() (types.Token, *types.ParseError) {
	tok := c.peek()
	if tok.Type != types.TokIdent {
		return tok, &types.ParseError{
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
