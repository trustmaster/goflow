package dsl

import "github.com/trustmaster/goflow"

// NewParser creates a parser graph that turns a Statement stream into a DefinitionResult.
//
//	In → RouteStatements → ParseExport  →
//	                    → ParseIIP      → CollectDefinition → Out
//	                    → ParseConnection →
func NewParser(f *goflow.Factory) (*goflow.Graph, error) {
	n := goflow.NewGraph()

	procs := []struct {
		name      string
		component string
	}{
		{"RouteStatements", "dsl/RouteStatements"},
		{"ParseExport", "dsl/ParseExport"},
		{"ParseIIP", "dsl/ParseIIP"},
		{"ParseConnection", "dsl/ParseConnection"},
		{"CollectDefinition", "dsl/CollectDefinition"},
	}

	for i := range procs {
		if err := n.AddNew(procs[i].name, procs[i].component, f); err != nil {
			return n, err
		}
	}

	conns := []struct{ src, srcPort, tgt, tgtPort string }{
		{"RouteStatements", "Export", "ParseExport", "In"},
		{"RouteStatements", "IIP", "ParseIIP", "In"},
		{"RouteStatements", "Connection", "ParseConnection", "In"},
		// Fan-in: all three parsers send fragments to CollectDefinition
		{"ParseExport", "Out", "CollectDefinition", "In"},
		{"ParseIIP", "Out", "CollectDefinition", "In"},
		{"ParseConnection", "Out", "CollectDefinition", "In"},
	}

	for i := range conns {
		c := conns[i]
		if err := n.Connect(c.src, c.srcPort, c.tgt, c.tgtPort); err != nil {
			return n, err
		}
	}

	n.MapInPort("In", "RouteStatements", "In")
	n.MapOutPort("Out", "CollectDefinition", "Out")

	return n, nil
}
