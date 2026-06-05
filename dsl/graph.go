// Package dsl implements FBP DSL parser.
package dsl

import (
	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/lex"
	"github.com/trustmaster/goflow/dsl/parse"
)

// RegisterComponents adds components of this library to the factory registry.
func RegisterComponents(f *goflow.Factory) error {
	if err := lex.RegisterLexerComponents(f); err != nil {
		return err
	}

	if err := parse.RegisterParseComponents(f); err != nil {
		return err
	}

	return nil
}
