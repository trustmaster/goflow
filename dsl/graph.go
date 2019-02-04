package dsl

import (
	"github.com/trustmaster/goflow"
)

// RegisterComponents adds components of this library to the factory registry
func RegisterComponents(f *goflow.Factory) error {
	if err := f.Register("dsl/ReadFile", func() (interface{}, error) {
		return new(ReadFile), nil
	}); err != nil {
		return err
	}
	return nil
}
