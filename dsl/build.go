package dsl

import (
	"fmt"

	"github.com/trustmaster/goflow"
)

// Build transforms a parsed Definition into a runnable *goflow.Graph.
// It creates a new graph, adds all processes, connects edges, adds IIPs,
// and maps exported ports. Returns a BuildError on any validation failure.
func Build(def *Definition, f *goflow.Factory) (*goflow.Graph, error) {
	g := goflow.NewGraph()

	// 1. Add all processes.
	//nolint:gocritic // map iteration always copies values in Go
	for name, proc := range def.Processes {
		if err := g.AddNew(name, proc.Component, f); err != nil {
			return nil, &BuildError{
				Err: fmt.Errorf("process %q: %w", name, err),
			}
		}
	}

	// 2. Connect all edges.
	if err := addConnections(g, def); err != nil {
		return nil, err
	}

	// 3. Add all IIPs.
	if err := addIIPs(g, def); err != nil {
		return nil, err
	}

	// 4. Map exported ports.
	if err := mapExports(g, def); err != nil {
		return nil, err
	}

	return g, nil
}

func addConnections(g *goflow.Graph, def *Definition) error {
	for i := range def.Connections {
		conn := def.Connections[i]

		if !processExists(def, conn.Src.Process) {
			return &BuildError{
				Err: fmt.Errorf("connect %s.%s -> %s.%s: source process %q not declared",
					conn.Src.Process, conn.Src.Port, conn.Tgt.Process, conn.Tgt.Port,
					conn.Src.Process),
			}
		}

		if !processExists(def, conn.Tgt.Process) {
			return &BuildError{
				Err: fmt.Errorf("connect %s.%s -> %s.%s: target process %q not declared",
					conn.Src.Process, conn.Src.Port, conn.Tgt.Process, conn.Tgt.Port,
					conn.Tgt.Process),
			}
		}

		srcPort := formatEndpointPort(conn.Src)
		tgtPort := formatEndpointPort(conn.Tgt)

		if err := g.Connect(conn.Src.Process, srcPort, conn.Tgt.Process, tgtPort); err != nil {
			return &BuildError{
				Err: fmt.Errorf("connect %s.%s -> %s.%s: %w",
					conn.Src.Process, srcPort, conn.Tgt.Process, tgtPort, err),
			}
		}
	}

	return nil
}

func addIIPs(g *goflow.Graph, def *Definition) error {
	for i := range def.IIPs {
		iip := def.IIPs[i]

		if !processExists(def, iip.Tgt.Process) {
			return &BuildError{
				Err: fmt.Errorf("iip -> %s.%s: target process %q not declared",
					iip.Tgt.Process, iip.Tgt.Port, iip.Tgt.Process),
			}
		}

		tgtPort := formatEndpointPort(iip.Tgt)

		if err := g.AddIIP(iip.Tgt.Process, tgtPort, iip.Data); err != nil {
			return &BuildError{
				Err: fmt.Errorf("iip -> %s.%s: %w", iip.Tgt.Process, tgtPort, err),
			}
		}
	}

	return nil
}

func mapExports(g *goflow.Graph, def *Definition) error {
	for i := range def.Exports {
		exp := def.Exports[i]

		if !processExists(def, exp.Proc) {
			return &BuildError{
				Err: fmt.Errorf("export %q references unknown process %q", exp.Public, exp.Proc),
			}
		}

		switch exp.Kind {
		case ExportInPort:
			g.MapInPort(exp.Public, exp.Proc, exp.Port)
		case ExportOutPort:
			g.MapOutPort(exp.Public, exp.Proc, exp.Port)
		default:
			return &BuildError{
				Err: fmt.Errorf("unknown export kind %q", exp.Kind),
			}
		}
	}

	return nil
}

// processExists checks whether a process name is declared in the Definition.
func processExists(def *Definition, name string) bool {
	_, ok := def.Processes[name]

	return ok
}

// formatEndpointPort formats an Endpoint's port name for use with goflow's
// Connect and AddIIP. The goflow library's parseAddress function handles the
// port[index] syntax for array ports.
func formatEndpointPort(ep Endpoint) string {
	if ep.Index != nil {
		return fmt.Sprintf("%s[%d]", ep.Port, *ep.Index)
	}

	return ep.Port
}
