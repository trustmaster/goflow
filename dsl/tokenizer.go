package dsl

import "github.com/trustmaster/goflow"

// NewTokenizer creates a Tokenizer graph.
func NewTokenizer(f *goflow.Factory) (*goflow.Graph, error) {
	n := goflow.NewGraph()

	if err := defineTokenizerProcs(n, f); err != nil {
		return n, err
	}

	if err := defineTokenizerConns(n); err != nil {
		return n, err
	}

	if err := defineTokenizerIIPs(n); err != nil {
		return n, err
	}

	n.MapInPort("In", "StartToken", "File")
	n.MapOutPort("Out", "Merge", "Out")

	return n, nil
}

func defineTokenizerProcs(n *goflow.Graph, f *goflow.Factory) error {
	procs := []struct {
		name      string
		component string
	}{
		{"StartToken", "dsl/StartToken"},
		{"Split", "dsl/Split"},
		{"Collect", "dsl/Collect"},
		{"Merge", "dsl/Merge"},
		{"ScanEOL", "dsl/ScanChars"},
		{"ScanWhitespace", "dsl/ScanChars"},
		{"ScanInt", "dsl/ScanChars"},
		{"ScanString", "dsl/ScanQuoted"},
		{"ScanEq", "dsl/ScanKeyword"},
		{"ScanDot", "dsl/ScanKeyword"},
		{"ScanColon", "dsl/ScanKeyword"},
		{"ScanLParen", "dsl/ScanKeyword"},
		{"ScanRParen", "dsl/ScanKeyword"},
		{"ScanArrow", "dsl/ScanKeyword"},
		{"ScanSlash", "dsl/ScanKeyword"},
		{"ScanHash", "dsl/ScanComment"},
		{"ScanInport", "dsl/ScanKeyword"},
		{"ScanOutport", "dsl/ScanKeyword"},
		{"ScanIdent", "dsl/ScanChars"},
	}

	for i := range procs {
		err := n.AddNew(procs[i].name, procs[i].component, f)
		if err != nil {
			return err
		}
	}

	return nil
}

func defineTokenizerConns(n *goflow.Graph) error {
	conns := []struct {
		srcName string
		srcPort string
		tgtName string
		tgtPort string
	}{
		{"StartToken", "Init", "Merge", ""},
		{"StartToken", "Next", "Split", ""},
		{"Collect", "Next", "Split", ""},
		{"Collect", "", "Merge", ""},
		{"Split", "Out[0]", "ScanEOL", ""},
		{"Split", "Out[1]", "ScanWhitespace", ""},
		{"Split", "Out[2]", "ScanInt", ""},
		{"Split", "Out[3]", "ScanString", ""},
		{"Split", "Out[4]", "ScanEq", ""},
		{"Split", "Out[5]", "ScanDot", ""},
		{"Split", "Out[6]", "ScanColon", ""},
		{"Split", "Out[7]", "ScanLParen", ""},
		{"Split", "Out[8]", "ScanRParen", ""},
		{"Split", "Out[9]", "ScanArrow", ""},
		{"Split", "Out[10]", "ScanSlash", ""},
		{"Split", "Out[11]", "ScanHash", ""},
		{"Split", "Out[12]", "ScanInport", ""},
		{"Split", "Out[13]", "ScanOutport", ""},
		{"Split", "Out[14]", "ScanIdent", ""},
		{"ScanEOL", "", "Collect", "In[0]"},
		{"ScanWhitespace", "", "Collect", "In[1]"},
		{"ScanInt", "", "Collect", "In[2]"},
		{"ScanString", "", "Collect", "In[3]"},
		{"ScanEq", "", "Collect", "In[4]"},
		{"ScanDot", "", "Collect", "In[5]"},
		{"ScanColon", "", "Collect", "In[6]"},
		{"ScanLParen", "", "Collect", "In[7]"},
		{"ScanRParen", "", "Collect", "In[8]"},
		{"ScanArrow", "", "Collect", "In[9]"},
		{"ScanSlash", "", "Collect", "In[10]"},
		{"ScanHash", "", "Collect", "In[11]"},
		{"ScanInport", "", "Collect", "In[12]"},
		{"ScanOutport", "", "Collect", "In[13]"},
		{"ScanIdent", "", "Collect", "In[14]"},
	}

	for i := range conns {
		if conns[i].srcPort == "" {
			conns[i].srcPort = "Out"
		}

		if conns[i].tgtPort == "" {
			conns[i].tgtPort = "In"
		}

		err := n.Connect(conns[i].srcName, conns[i].srcPort, conns[i].tgtName, conns[i].tgtPort)
		if err != nil {
			return err
		}
	}

	return nil
}

func defineTokenizerIIPs(n *goflow.Graph) error {
	iips := []struct {
		proc, port string
		val        TokenType
	}{
		{"ScanEOL", "SET", "\r\n"},
		{"ScanEOL", "TYPE", tokEOL},
		{"ScanWhitespace", "SET", "\t "},
		{"ScanWhitespace", "TYPE", tokWhitespace},
		{"ScanInt", "SET", "0123456789"},
		{"ScanInt", "TYPE", tokInt},
		{"ScanString", "SET", "\"'"},
		{"ScanString", "TYPE", tokQuotedStr},
		{"ScanEq", "SET", "="},
		{"ScanEq", "TYPE", tokEqual},
		{"ScanDot", "SET", "."},
		{"ScanDot", "TYPE", tokDot},
		{"ScanColon", "SET", ":"},
		{"ScanColon", "TYPE", tokColon},
		{"ScanLParen", "SET", "("},
		{"ScanLParen", "TYPE", tokLparen},
		{"ScanRParen", "SET", ")"},
		{"ScanRParen", "TYPE", tokRparen},
		{"ScanArrow", "SET", "->"},
		{"ScanArrow", "TYPE", tokArrow},
		{"ScanSlash", "SET", "/"},
		{"ScanSlash", "TYPE", tokSlash},
		{"ScanHash", "SET", "#"},
		{"ScanHash", "TYPE", tokComment},
		{"ScanInport", "SET", "INPORT"},
		{"ScanInport", "TYPE", tokInport},
		{"ScanOutport", "SET", "OUTPORT"},
		{"ScanOutport", "TYPE", tokOutport},
		{"ScanIdent", "SET", "[\\w_]"},
		{"ScanIdent", "TYPE", tokIdent},
	}

	for i := range iips {
		err := n.AddIIP(iips[i].proc, iips[i].port, string(iips[i].val))
		if err != nil {
			return err
		}
	}

	return nil
}
