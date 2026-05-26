package dsl

import "github.com/trustmaster/goflow"

// NewLexer creates a deterministic lexer graph.
func NewLexer(f *goflow.Factory) (*goflow.Graph, error) {
	n := goflow.NewGraph()

	if err := defineLexerProcs(n, f); err != nil {
		return n, err
	}

	if err := defineLexerConns(n); err != nil {
		return n, err
	}

	n.MapInPort("In", "StartCursor", "File")
	n.MapOutPort("Out", "Advance", "Out")

	return n, nil
}

func defineLexerProcs(n *goflow.Graph, f *goflow.Factory) error {
	procs := []struct {
		name      string
		component string
	}{
		{"StartCursor", "dsl/StartCursor"},
		{"Dispatch", "dsl/Dispatch"},
		{"ScanWhitespace", "dsl/ScanWhitespaceToken"},
		{"ScanComment", "dsl/ScanCommentToken"},
		{"ScanQuoted", "dsl/ScanQuotedToken"},
		{"ScanIdent", "dsl/ScanIdentToken"},
		{"ScanNumber", "dsl/ScanNumberToken"},
		{"ScanOperator", "dsl/ScanOperatorToken"},
		{"Advance", "dsl/Advance"},
	}

	for i := range procs {
		if err := n.AddNew(procs[i].name, procs[i].component, f); err != nil {
			return err
		}
	}

	return nil
}

func defineLexerConns(n *goflow.Graph) error {
	conns := []struct {
		srcName string
		srcPort string
		tgtName string
		tgtPort string
	}{
		{"StartCursor", "Out", "Dispatch", "In"},
		{"Advance", "Next", "Dispatch", "In"},
		{"Dispatch", "Whitespace", "ScanWhitespace", "In"},
		{"Dispatch", "Comment", "ScanComment", "In"},
		{"Dispatch", "Quoted", "ScanQuoted", "In"},
		{"Dispatch", "Ident", "ScanIdent", "In"},
		{"Dispatch", "Number", "ScanNumber", "In"},
		{"Dispatch", "Operator", "ScanOperator", "In"},
		{"Dispatch", "Eof", "Advance", "In"},
		{"ScanWhitespace", "Out", "Advance", "In"},
		{"ScanComment", "Out", "Advance", "In"},
		{"ScanQuoted", "Out", "Advance", "In"},
		{"ScanIdent", "Out", "Advance", "In"},
		{"ScanNumber", "Out", "Advance", "In"},
		{"ScanOperator", "Out", "Advance", "In"},
	}

	for i := range conns {
		if err := n.Connect(conns[i].srcName, conns[i].srcPort, conns[i].tgtName, conns[i].tgtPort); err != nil {
			return err
		}
	}

	return nil
}
