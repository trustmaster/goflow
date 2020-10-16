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

// RegisterComponents adds components of this library to the factory registry
func RegisterComponents(f *goflow.Factory) error {
	return registerComponentConstructors(f, []componentConstructor{
		{"dsl/Collect", func() (interface{}, error) {
			return new(Collect), nil
		}},
		{"dsl/Merge", func() (interface{}, error) {
			return new(Merge), nil
		}},
		{"dsl/Reader", func() (interface{}, error) {
			return new(Reader), nil
		}},
		{"dsl/ScanChars", func() (interface{}, error) {
			return new(ScanChars), nil
		}},
		{"dsl/ScanComment", func() (interface{}, error) {
			return new(ScanComment), nil
		}},
		{"dsl/ScanKeyword", func() (interface{}, error) {
			return new(ScanKeyword), nil
		}},
		{"dsl/ScanQuoted", func() (interface{}, error) {
			return new(ScanQuoted), nil
		}},
		{"dsl/Split", func() (interface{}, error) {
			return new(Split), nil
		}},
		{"dsl/StartToken", func() (interface{}, error) {
			return new(StartToken), nil
		}},
		{"dsl/Tokenizer", func() (interface{}, error) {
			return NewTokenizer(f)
		}},
	})
}
