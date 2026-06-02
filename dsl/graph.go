// Package dsl implements FBP DSL parser.
package dsl

import (
	"github.com/trustmaster/goflow"
)

type componentConstructor struct {
	name string
	ctor func() (interface{}, error)
}

func registerComponentConstructors(f *goflow.Factory, list []componentConstructor) error {
	for i := range list {
		err := f.Register(list[i].name, list[i].ctor)
		if err != nil {
			return err
		}
	}

	return nil
}

// RegisterComponents adds components of this library to the factory registry.
func RegisterComponents(f *goflow.Factory) error {
	return registerComponentConstructors(f, []componentConstructor{
		{"dsl/Advance", func() (interface{}, error) {
			return new(Advance), nil
		}},
		{"dsl/Dispatch", func() (interface{}, error) {
			return new(Dispatch), nil
		}},
		{"dsl/Lexer", func() (interface{}, error) {
			return NewLexer(f)
		}},
		{"dsl/Reader", func() (interface{}, error) {
			return new(Reader), nil
		}},
		{"dsl/SegmentStatements", func() (interface{}, error) {
			return new(SegmentStatements), nil
		}},
		{"dsl/ScanCommentToken", func() (interface{}, error) {
			return new(ScanCommentToken), nil
		}},
		{"dsl/ScanIdentToken", func() (interface{}, error) {
			return new(ScanIdentToken), nil
		}},
		{"dsl/ScanNumberToken", func() (interface{}, error) {
			return new(ScanNumberToken), nil
		}},
		{"dsl/ScanOperatorToken", func() (interface{}, error) {
			return new(ScanOperatorToken), nil
		}},
		{"dsl/ScanQuotedToken", func() (interface{}, error) {
			return new(ScanQuotedToken), nil
		}},
		{"dsl/ScanWhitespaceToken", func() (interface{}, error) {
			return new(ScanWhitespaceToken), nil
		}},
		{"dsl/StripTrivia", func() (interface{}, error) {
			return new(StripTrivia), nil
		}},
		{"dsl/StartCursor", func() (interface{}, error) {
			return new(StartCursor), nil
		}},
		{"dsl/RouteStatements", func() (interface{}, error) {
			return new(RouteStatements), nil
		}},
		{"dsl/ParseExport", func() (interface{}, error) {
			return new(ParseExport), nil
		}},
		{"dsl/ParseIIP", func() (interface{}, error) {
			return new(ParseIIP), nil
		}},
		{"dsl/ParseConnection", func() (interface{}, error) {
			return new(ParseConnection), nil
		}},
		{"dsl/CollectDefinition", func() (interface{}, error) {
			return new(CollectDefinition), nil
		}},
		{"dsl/Parser", func() (interface{}, error) {
			return NewParser(f)
		}},
	})
}
