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
		{"dsl/Collect", func() (interface{}, error) {
			return new(Collect), nil
		}},
		{"dsl/Dispatch", func() (interface{}, error) {
			return new(Dispatch), nil
		}},
		{"dsl/Lexer", func() (interface{}, error) {
			return NewLexer(f)
		}},
		{"dsl/Merge", func() (interface{}, error) {
			return new(Merge), nil
		}},
		{"dsl/Reader", func() (interface{}, error) {
			return new(Reader), nil
		}},
		{"dsl/SegmentStatements", func() (interface{}, error) {
			return new(SegmentStatements), nil
		}},
		{"dsl/ScanChars", func() (interface{}, error) {
			return new(ScanChars), nil
		}},
		{"dsl/ScanComment", func() (interface{}, error) {
			return new(ScanComment), nil
		}},
		{"dsl/ScanCommentToken", func() (interface{}, error) {
			return new(ScanCommentToken), nil
		}},
		{"dsl/ScanIdentToken", func() (interface{}, error) {
			return new(ScanIdentToken), nil
		}},
		{"dsl/ScanKeyword", func() (interface{}, error) {
			return new(ScanKeyword), nil
		}},
		{"dsl/ScanNumberToken", func() (interface{}, error) {
			return new(ScanNumberToken), nil
		}},
		{"dsl/ScanOperatorToken", func() (interface{}, error) {
			return new(ScanOperatorToken), nil
		}},
		{"dsl/ScanQuoted", func() (interface{}, error) {
			return new(ScanQuoted), nil
		}},
		{"dsl/ScanQuotedToken", func() (interface{}, error) {
			return new(ScanQuotedToken), nil
		}},
		{"dsl/ScanWhitespaceToken", func() (interface{}, error) {
			return new(ScanWhitespaceToken), nil
		}},
		{"dsl/Split", func() (interface{}, error) {
			return new(Split), nil
		}},
		{"dsl/StripTrivia", func() (interface{}, error) {
			return new(StripTrivia), nil
		}},
		{"dsl/StartCursor", func() (interface{}, error) {
			return new(StartCursor), nil
		}},
		{"dsl/StartToken", func() (interface{}, error) {
			return new(StartToken), nil
		}},
		{"dsl/Tokenizer", func() (interface{}, error) {
			return NewTokenizer(f)
		}},
	})
}
