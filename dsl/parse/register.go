package parse

import "github.com/trustmaster/goflow"

// RegisterParseComponents registers all parse components with the factory.
func RegisterParseComponents(f *goflow.Factory) error {
	components := map[string]func() (interface{}, error){
		"dsl/Parser": func() (interface{}, error) {
			return NewParser(f)
		},
		"dsl/Reader": func() (interface{}, error) {
			return new(Reader), nil
		},
		"dsl/StripTrivia": func() (interface{}, error) {
			return new(StripTrivia), nil
		},
		"dsl/SegmentStatements": func() (interface{}, error) {
			return new(SegmentStatements), nil
		},
		"dsl/RouteStatements": func() (interface{}, error) {
			return new(RouteStatements), nil
		},
		"dsl/ParseExport": func() (interface{}, error) {
			return new(ParseExport), nil
		},
		"dsl/ParseIIP": func() (interface{}, error) {
			return new(ParseIIP), nil
		},
		"dsl/ParseConnection": func() (interface{}, error) {
			return new(ParseConnection), nil
		},
		"dsl/CollectDefinition": func() (interface{}, error) {
			return new(CollectDefinition), nil
		},
	}

	for name, ctor := range components {
		if err := f.Register(name, ctor); err != nil {
			return err
		}
	}

	return nil
}
