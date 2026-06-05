package parse

import "github.com/trustmaster/goflow"

// NewParser creates a parser graph that turns a Statement stream into a DefinitionResult.
//
//	In → RouteStatements → ParseExport  →
//	                    → ParseIIP      → CollectDefinition → Out
//	                    → ParseConnection →
func NewParser(f *goflow.Factory) (*goflow.Graph, error) {
	n := goflow.NewGraph()

	const (
		procRoute       = "RouteStatements"
		procParseExport = "ParseExport"
		procParseIIP    = "ParseIIP"
		procParseConn   = "ParseConnection"
		procCollectDef  = "CollectDefinition"
		compRoute       = "dsl/RouteStatements"
		compParseExport = "dsl/ParseExport"
		compParseIIP    = "dsl/ParseIIP"
		compParseConn   = "dsl/ParseConnection"
		compCollectDef  = "dsl/CollectDefinition"
	)

	procs := []struct {
		name      string
		component string
	}{
		{procRoute, compRoute},
		{procParseExport, compParseExport},
		{procParseIIP, compParseIIP},
		{procParseConn, compParseConn},
		{procCollectDef, compCollectDef},
	}

	for i := range procs {
		if err := n.AddNew(procs[i].name, procs[i].component, f); err != nil {
			return n, err
		}
	}

	//nolint:goconst // "Out" and "In" are common port names, not configurable constants
	conns := []struct{ src, srcPort, tgt, tgtPort string }{
		{procRoute, "Export", procParseExport, "In"},
		{procRoute, "IIP", procParseIIP, "In"},
		{procRoute, "Connection", procParseConn, "In"},
		// Fan-in: all three parsers send fragments to CollectDefinition
		{procParseExport, "Out", procCollectDef, "In"},
		{procParseIIP, "Out", procCollectDef, "In"},
		{procParseConn, "Out", procCollectDef, "In"},
	}

	for i := range conns {
		c := conns[i]

		if err := n.Connect(c.src, c.srcPort, c.tgt, c.tgtPort); err != nil {
			return n, err
		}
	}

	n.MapInPort("In", procRoute, "In")
	n.MapOutPort("Out", procCollectDef, "Out")

	return n, nil
}
