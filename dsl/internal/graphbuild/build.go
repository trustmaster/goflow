package graphbuild

import (
	"errors"
	"fmt"

	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/types"
)

// Build transforms a parsed Definition into a runnable *goflow.Graph.
// It creates a new graph, adds all processes, connects edges, adds IIPs,
// and maps exported ports. Returns a BuildError on any validation failure.
func Build(def *types.Definition, f *goflow.Factory) (*goflow.Graph, error) {
	if def == nil {
		return nil, &types.BuildError{Err: errors.New("definition cannot be nil")}
	}

	if f == nil {
		return nil, &types.BuildError{Err: errors.New("factory cannot be nil")}
	}

	g := goflow.NewGraph()

	for name := range def.Processes {
		if err := g.AddNew(name, def.Processes[name].Component, f); err != nil {
			return nil, &types.BuildError{
				Err: fmt.Errorf("process %q: %w", name, err),
			}
		}
	}

	if err := addConnections(g, def); err != nil {
		return nil, err
	}

	if err := addIIPs(g, def); err != nil {
		return nil, err
	}

	if err := mapExports(g, def); err != nil {
		return nil, err
	}

	return g, nil
}

func addConnections(g *goflow.Graph, def *types.Definition) error {
	for i := range def.Connections {
		conn := def.Connections[i]

		if !processExists(def, conn.Src.Process) {
			return &types.BuildError{
				Err: fmt.Errorf("connect %s.%s -> %s.%s: source process %q not declared",
					conn.Src.Process, conn.Src.Port, conn.Tgt.Process, conn.Tgt.Port,
					conn.Src.Process),
			}
		}

		if !processExists(def, conn.Tgt.Process) {
			return &types.BuildError{
				Err: fmt.Errorf("connect %s.%s -> %s.%s: target process %q not declared",
					conn.Src.Process, conn.Src.Port, conn.Tgt.Process, conn.Tgt.Port,
					conn.Tgt.Process),
			}
		}

		srcPort := formatEndpointPort(conn.Src)
		tgtPort := formatEndpointPort(conn.Tgt)

		if err := g.Connect(conn.Src.Process, srcPort, conn.Tgt.Process, tgtPort); err != nil {
			return &types.BuildError{
				Err: fmt.Errorf("connect %s.%s -> %s.%s: %w",
					conn.Src.Process, srcPort, conn.Tgt.Process, tgtPort, err),
			}
		}
	}

	return nil
}

func addIIPs(g *goflow.Graph, def *types.Definition) error {
	for i := range def.IIPs {
		iip := def.IIPs[i]

		if !processExists(def, iip.Tgt.Process) {
			return &types.BuildError{
				Err: fmt.Errorf("iip -> %s.%s: target process %q not declared",
					iip.Tgt.Process, iip.Tgt.Port, iip.Tgt.Process),
			}
		}

		tgtPort := formatEndpointPort(iip.Tgt)

		if err := g.AddIIP(iip.Tgt.Process, tgtPort, iip.Data); err != nil {
			return &types.BuildError{
				Err: fmt.Errorf("iip -> %s.%s: %w", iip.Tgt.Process, tgtPort, err),
			}
		}
	}

	return nil
}

func mapExports(g *goflow.Graph, def *types.Definition) error {
	for i := range def.Exports {
		exp := def.Exports[i]

		if !processExists(def, exp.Proc) {
			return &types.BuildError{
				Err: fmt.Errorf("export %q references unknown process %q", exp.Public, exp.Proc),
			}
		}

		switch exp.Kind {
		case types.ExportInPort:
			g.MapInPort(exp.Public, exp.Proc, exp.Port)
		case types.ExportOutPort:
			g.MapOutPort(exp.Public, exp.Proc, exp.Port)
		default:
			return &types.BuildError{
				Err: fmt.Errorf("unknown export kind %q", exp.Kind),
			}
		}
	}

	return nil
}

func processExists(def *types.Definition, name string) bool {
	_, ok := def.Processes[name]
	return ok
}

func formatEndpointPort(ep types.Endpoint) string {
	if ep.Index != nil {
		return fmt.Sprintf("%s[%d]", ep.Port, *ep.Index)
	}

	return ep.Port
}
