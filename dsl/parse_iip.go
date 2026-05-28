package dsl

import "fmt"

// ParseIIP is a FBP component that parses IIP (initial information packet) statements.
type ParseIIP struct {
	In  <-chan Statement
	Out chan<- Fragment
}

// Process parses each incoming Statement as an IIP declaration.
func (p *ParseIIP) Process() {
	for stmt := range p.In {
		frags, err := parseIIPStatement(stmt)
		if err != nil {
			p.Out <- Fragment{Kind: FragmentError, Err: err}
			continue
		}

		for i := range frags {
			p.Out <- frags[i]
		}
	}
}

// parseIIPStatement parses a single IIP statement.
//
// Expected token sequence:
//
//	'data'  -> PortName [[Index]] ProcName [( ComponentPath )]
//	42      -> PortName [[Index]] ProcName [( ComponentPath )]
//
// Quoted strings have outer quote characters stripped for the Data field.
// Integer values are converted to int.
func parseIIPStatement(stmt Statement) ([]Fragment, *ParseError) {
	cur := newTokenCursor(stmt.Tokens)

	dataTok := cur.consume()

	var data any

	switch dataTok.Type {
	case TokQuoted:
		v := dataTok.Value
		if len(v) >= 2 {
			v = v[1 : len(v)-1] // strip surrounding quote characters
		}

		data = v
	case TokInt:
		n := 0
		for _, ch := range dataTok.Value {
			if ch >= '0' && ch <= '9' {
				n = n*10 + int(ch-'0')
			}
		}

		data = n
	default:
		return nil, &ParseError{
			Span: dataTok.Span,
			Err:  fmt.Errorf("expected quoted string or integer for IIP data, got %s %q", dataTok.Type, dataTok.Value),
		}
	}

	if _, err := cur.expect(TokArrow); err != nil {
		return nil, err
	}

	tgtPortTok, err := cur.expectIdent()
	if err != nil {
		return nil, err
	}

	tgtIndex, err := parseArrayIndex(cur)
	if err != nil {
		return nil, err
	}

	tgtProcTok, err := cur.expectIdent()
	if err != nil {
		return nil, err
	}

	tgtComp, err := parseComponent(cur)
	if err != nil {
		return nil, err
	}

	var frags []Fragment

	if tgtComp != "" {
		frags = append(frags, Fragment{
			Kind:    FragmentProcess,
			Process: &ProcessDef{Name: tgtProcTok.Value, Component: tgtComp},
		})
	}

	frags = append(frags, Fragment{
		Kind: FragmentIIP,
		IIP: &IIPDef{
			Data: data,
			Tgt:  Endpoint{Process: tgtProcTok.Value, Port: tgtPortTok.Value, Index: tgtIndex},
		},
	})

	return frags, nil
}
