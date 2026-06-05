package lex

import "github.com/trustmaster/goflow"

// RegisterLexerComponents registers all lexer components with the factory.
func RegisterLexerComponents(f *goflow.Factory) error {
	components := map[string]func() (interface{}, error){
		"dsl/Lexer": func() (interface{}, error) {
			return NewLexer(f)
		},
		"dsl/StartCursor": func() (interface{}, error) {
			return new(StartCursor), nil
		},
		"dsl/Dispatch": func() (interface{}, error) {
			return new(Dispatch), nil
		},
		"dsl/ScanWhitespaceToken": func() (interface{}, error) {
			return new(ScanWhitespaceToken), nil
		},
		"dsl/ScanCommentToken": func() (interface{}, error) {
			return new(ScanCommentToken), nil
		},
		"dsl/ScanQuotedToken": func() (interface{}, error) {
			return new(ScanQuotedToken), nil
		},
		"dsl/ScanIdentToken": func() (interface{}, error) {
			return new(ScanIdentToken), nil
		},
		"dsl/ScanNumberToken": func() (interface{}, error) {
			return new(ScanNumberToken), nil
		},
		"dsl/ScanOperatorToken": func() (interface{}, error) {
			return new(ScanOperatorToken), nil
		},
		"dsl/Advance": func() (interface{}, error) {
			return new(Advance), nil
		},
	}

	for name, ctor := range components {
		if err := f.Register(name, ctor); err != nil {
			return err
		}
	}

	return nil
}
