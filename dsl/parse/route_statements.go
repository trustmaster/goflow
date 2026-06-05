package parse

import "github.com/trustmaster/goflow/dsl/types"

// RouteStatements classifies statements and routes each one to the appropriate output port.
// Export statements (INPORT/OUTPORT) go to Export.
// IIP statements (quoted string or integer followed by ->) go to IIP.
// All other statements go to Connection.
type RouteStatements struct {
	In         <-chan types.Statement
	Export     chan<- types.Statement
	Iip        chan<- types.Statement
	Connection chan<- types.Statement
}

// Process dispatches each incoming statement to exactly one output.
func (r *RouteStatements) Process() {
	for stmt := range r.In {
		if len(stmt.Tokens) == 0 {
			continue
		}

		first := stmt.Tokens[0]

		switch {
		case first.Type == types.TokIdent && (first.Value == "INPORT" || first.Value == "OUTPORT"):
			r.Export <- stmt
		case first.Type == types.TokQuoted || first.Type == types.TokInt:
			r.Iip <- stmt
		default:
			r.Connection <- stmt
		}
	}
}
