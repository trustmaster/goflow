package lex

import (
	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/internal/graphbuild"
)

// NewLexer creates the lexer graph from the generated internal definition.
func NewLexer(f *goflow.Factory) (*goflow.Graph, error) {
	return graphbuild.Build(&generatedLexerDefinition, f)
}
