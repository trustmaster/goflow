package dsl

// RouteStatements classifies statements and routes each one to the appropriate output port.
// Export statements (INPORT/OUTPORT) go to Export.
// IIP statements (quoted string or integer followed by ->) go to IIP.
// All other statements go to Connection.
type RouteStatements struct {
	In         <-chan Statement
	Export     chan<- Statement
	Iip        chan<- Statement
	Connection chan<- Statement
}

// Process dispatches each incoming statement to exactly one output.
func (r *RouteStatements) Process() {
	for stmt := range r.In {
		if len(stmt.Tokens) == 0 {
			continue
		}

		first := stmt.Tokens[0]

		switch {
		case first.Type == TokIdent && (first.Value == keywordINPORT || first.Value == keywordOUTPORT):
			r.Export <- stmt
		case first.Type == TokQuoted || first.Type == TokInt:
			r.Iip <- stmt
		default:
			r.Connection <- stmt
		}
	}
}
