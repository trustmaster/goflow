package types

import "fmt"

// LexError is a lexical error.
type LexError struct {
	Span Span
	Err  error
}

// Error returns an error message.
func (e LexError) Error() string {
	return fmt.Sprintf("lex error at %s:%d:%d: %s", e.Span.File, e.Span.Line, e.Span.Column, e.Err)
}

// ParseError is a parser error.
type ParseError struct {
	Span Span
	Err  error
}

// Error returns an error message.
func (e ParseError) Error() string {
	return fmt.Sprintf("parse error at %s:%d:%d: %s", e.Span.File, e.Span.Line, e.Span.Column, e.Err)
}

// BuildError is an error while turning a parsed definition into a graph.
type BuildError struct {
	Err error
}

// Error returns an error message.
func (e BuildError) Error() string {
	return fmt.Sprintf("build error: %s", e.Err)
}
