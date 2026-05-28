package dsl

import "fmt"

// CollectDefinition accumulates parsed fragments into a single DefinitionResult.
// It reads until its input is closed, then emits one DefinitionResult.
type CollectDefinition struct {
	In  <-chan Fragment
	Out chan<- DefinitionResult
}

// Process reads all fragments and emits the accumulated DefinitionResult.
func (c *CollectDefinition) Process() {
	result := DefinitionResult{
		Definition: Definition{
			Processes: make(map[string]ProcessDef),
		},
	}

	for frag := range c.In {
		switch frag.Kind {
		case FragmentProcess:
			if err := mergeProcess(&result.Definition, frag.Process); err != nil {
				result.Errors = append(result.Errors, err)
			}
		case FragmentConnection:
			result.Definition.Connections = append(result.Definition.Connections, *frag.Connection)
		case FragmentIIP:
			result.Definition.IIPs = append(result.Definition.IIPs, *frag.IIP)
		case FragmentExport:
			result.Definition.Exports = append(result.Definition.Exports, *frag.Export)
		case FragmentError:
			result.Errors = append(result.Errors, frag.Err)
		}
	}

	c.Out <- result
}

// mergeProcess inserts or validates a process declaration.
// Consistent duplicates (same name and same component) are accepted.
// Conflicting declarations (same name, different component) produce a BuildError.
func mergeProcess(def *Definition, proc *ProcessDef) error {
	existing, ok := def.Processes[proc.Name]
	if !ok {
		def.Processes[proc.Name] = *proc
		return nil
	}

	if existing.Component != proc.Component {
		return &BuildError{
			Err: fmt.Errorf("conflicting component for process %q: %q vs %q",
				proc.Name, existing.Component, proc.Component),
		}
	}

	return nil // consistent duplicate — ok
}
