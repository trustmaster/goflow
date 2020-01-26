package dsl

import (
	"github.com/trustmaster/goflow"
)

// RegisterComponents adds components of this library to the factory registry
func RegisterComponents(f *goflow.Factory) error {
	if err := f.Register("dsl/Reader", func() (interface{}, error) {
		return new(Reader), nil
	}); err != nil {
		return err
	}
	if err := f.Register("dsl/Tokenizer", func() (interface{}, error) {
		return new(Tokenizer), nil
	}); err != nil {
		return err
	}
	if err := f.Register("dsl/ScanChars", func() (interface{}, error) {
		return new(ScanChars), nil
	}); err != nil {
		return err
	}
	if err := f.Register("dsl/ScanKeyword", func() (interface{}, error) {
		return new(ScanKeyword), nil
	}); err != nil {
		return err
	}
	if err := f.Register("dsl/ScanComment", func() (interface{}, error) {
		return new(ScanComment), nil
	}); err != nil {
		return err
	}
	return nil
}
