package parse

import "github.com/trustmaster/goflow"

// NewParser creates the parser graph from the generated internal definition.
func NewParser(f *goflow.Factory) (*goflow.Graph, error) {
	return Build(&generatedParserDefinition, f)
}
