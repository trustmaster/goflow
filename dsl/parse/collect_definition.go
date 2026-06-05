package parse

import (
	"fmt"

	"github.com/trustmaster/goflow/dsl/types"
)

// CollectDefinition accumulates parsed fragments into a single DefinitionResult.
// It reads until its input is closed, then emits one DefinitionResult.
type CollectDefinition struct {
	In  <-chan types.Fragment
	Out chan<- types.DefinitionResult
}

// Process reads all fragments and emits the accumulated DefinitionResult.
func (c *CollectDefinition) Process() {
	result := types.DefinitionResult{
		Definition: types.Definition{
			Processes: make(map[string]types.ProcessDef),
		},
	}

	for frag := range c.In {
		switch frag.Kind {
		case types.FragmentProcess:
			if err := mergeProcess(&result.Definition, frag.Process); err != nil {
				result.Errors = append(result.Errors, err)
			}
		case types.FragmentConnection:
			result.Definition.Connections = append(result.Definition.Connections, *frag.Connection)
		case types.FragmentIIP:
			result.Definition.IIPs = append(result.Definition.IIPs, *frag.IIP)
		case types.FragmentExport:
			result.Definition.Exports = append(result.Definition.Exports, *frag.Export)
		case types.FragmentError:
			result.Errors = append(result.Errors, frag.Err)
		}
	}

	c.Out <- result
}

// mergeProcess inserts or validates a process declaration.
// Consistent duplicates (same name and same component) are accepted.
// Conflicting declarations (same name, different component) produce a BuildError.
func mergeProcess(def *types.Definition, proc *types.ProcessDef) error {
	existing, ok := def.Processes[proc.Name]
	if !ok {
		def.Processes[proc.Name] = *proc
		return nil
	}

	if existing.Component != proc.Component {
		return &types.BuildError{
			Err: fmt.Errorf("conflicting component for process %q: %q vs %q",
				proc.Name, existing.Component, proc.Component),
		}
	}

	return nil // consistent duplicate — ok
}
