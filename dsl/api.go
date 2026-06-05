package dsl

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/parse"
	"github.com/trustmaster/goflow/dsl/types"
)

// ParseDefinition parses FBP source bytes into a Definition.
func ParseDefinition(src []byte) (*types.Definition, error) {
	if len(src) == 0 {
		return &types.Definition{
			Processes:   make(map[string]types.ProcessDef),
			Connections: []types.ConnectionDef{},
			IIPs:        []types.IIPDef{},
			Exports:     []types.ExportDef{},
		}, nil
	}

	return parseFile(&types.File{Name: "<input>", Data: src})
}

// LoadDefinitionFile reads an FBP file from disk and parses it into a Definition.
func LoadDefinitionFile(path string) (*types.Definition, error) {
	data, err := os.ReadFile(filepath.Clean(path)) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("load definition file: %w", err)
	}

	return parseFile(&types.File{Name: path, Data: data})
}

// Parse parses FBP source bytes and builds a runnable *goflow.Graph using the
// provided component factory.
func Parse(src []byte, f *goflow.Factory) (*goflow.Graph, error) {
	def, err := ParseDefinition(src)
	if err != nil {
		return nil, err
	}

	return parse.Build(def, f)
}

// LoadFile reads an FBP file from disk and builds a runnable *goflow.Graph
// using the provided component factory.
func LoadFile(path string, f *goflow.Factory) (*goflow.Graph, error) {
	def, err := LoadDefinitionFile(path)
	if err != nil {
		return nil, err
	}

	return parse.Build(def, f)
}

// UnmarshalDefinition unmarshals a JSON-encoded Definition.
func UnmarshalDefinition(data []byte) (*types.Definition, error) {
	var def types.Definition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, err
	}

	return &def, nil
}

// parseFile runs the internal top-level DSL pipeline graph and returns the collected Definition.
func parseFile(file *types.File) (*types.Definition, error) {
	result, err := runInternalPipeline(file)
	if err != nil {
		return nil, err
	}

	if len(result.Errors) > 0 {
		return nil, errors.Join(result.Errors...)
	}

	return &result.Definition, nil
}

// Build builds a runnable *goflow.Graph from a Definition.
// This is a convenience wrapper around parse.Build.
func Build(def *types.Definition, f *goflow.Factory) (*goflow.Graph, error) {
	return parse.Build(def, f)
}
