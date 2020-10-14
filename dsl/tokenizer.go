package dsl

import (
	"fmt"

	"github.com/trustmaster/goflow"
)

// TokenType differentiates tokens
type TokenType string

const (
	// Token types
	tokIllegal    = TokenType("illegal")    // S:Fallback P:9
	tokNewFile    = TokenType("newFile")    // S:Auto
	tokEOF        = TokenType("eof")        // S:Auto
	tokWhitespace = TokenType("whitespace") // S:Chars P:1
	tokEOL        = TokenType("eol")        // S:Chars P:1
	tokComment    = TokenType("comment")    // # S:Comment P:2

	// Literals
	tokIdent     = TokenType("ident")     // ProcName S:Chars P:3
	tokInt       = TokenType("int")       // 123 S:Chars P:2
	tokQuotedStr = TokenType("quotedStr") // "data" S:Quoted P:2

	// Operators
	tokEqual  = TokenType("equal")  // = S:Keyword P:2
	tokDot    = TokenType("dot")    // . S:Keyword P:2
	tokColon  = TokenType("colon")  // : S:Keyword P:2
	tokLparen = TokenType("lparen") // ( S:Keyword P:2
	tokRparen = TokenType("rparen") // ) S:Keyword P:2
	tokArrow  = TokenType("arrow")  //", "S:Keyword P:2
	tokSlash  = TokenType("slash")  // / S:Keyword P:2

	// Keywords
	tokInport  = TokenType("inport")  // S:Keyword P:2
	tokOutport = TokenType("outport") // S:Keyword P:2
)

// Token represents a single lexem in a File
type Token struct {
	Type  TokenType
	File  *File
	Pos   int
	Value string
}

func (t Token) String() string {
	return t.Value
}

// LexError is a lexical error
type LexError struct {
	File string
	Pos  int
	Err  error
}

// Error returns an error message
func (e LexError) Error() string {
	return fmt.Sprintf("Error scanning file '%s' at pos %d: %s", e.File, e.Pos, e.Err.Error())
}

// NewTokenizer creates a Tokenizer graph
func NewTokenizer(f *goflow.Factory) (*goflow.Graph, error) {
	n := goflow.NewGraph()

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

	for _, p := range procs {
		err := n.AddNew(p.name, p.component, f)
		if err != nil {
			return n, err
		}
	}

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

	for _, c := range conns {
		if c.srcPort == "" {
			c.srcPort = "Out"
		}
		if c.tgtPort == "" {
			c.tgtPort = "In"
		}
		err := n.Connect(c.srcName, c.srcPort, c.tgtName, c.tgtPort)
		if err != nil {
			return n, err
		}
	}

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

	for _, iip := range iips {
		err := n.AddIIP(iip.proc, iip.port, string(iip.val))
		if err != nil {
			return n, err
		}
	}

	n.MapInPort("In", "StartToken", "File")
	n.MapOutPort("Out", "Merge", "Out")

	return n, nil
}
