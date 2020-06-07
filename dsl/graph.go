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
		return NewTokenizer(f)
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
	if err := f.Register("dsl/ScanQuoted", func() (interface{}, error) {
		return new(ScanQuoted), nil
	}); err != nil {
		return err
	}
	if err := f.Register("dsl/Split", func() (interface{}, error) {
		return new(Split), nil
	}); err != nil {
		return err
	}
	if err := f.Register("dsl/Collect", func() (interface{}, error) {
		return new(Collect), nil
	}); err != nil {
		return err
	}
	if err := f.Register("dsl/StartToken", func() (interface{}, error) {
		return new(StartToken), nil
	}); err != nil {
		return err
	}
	if err := f.Register("dsl/Merge", func() (interface{}, error) {
		return new(Merge), nil
	}); err != nil {
		return err
	}
	return nil
}
