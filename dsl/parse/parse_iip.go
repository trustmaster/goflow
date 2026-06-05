package parse

import (
	"fmt"
	"strconv"

	"github.com/trustmaster/goflow/dsl/types"
)

// ParseIIP is a FBP component that parses IIP (initial information packet) statements.
type ParseIIP struct {
	In  <-chan types.Statement
	Out chan<- types.Fragment
}

// Process parses each incoming Statement as an IIP declaration.
func (p *ParseIIP) Process() {
	for stmt := range p.In {
		frags, err := parseIIPStatement(stmt)
		if err != nil {
			p.Out <- types.Fragment{Kind: types.FragmentError, Err: err}
			continue
		}

		for i := range frags {
			p.Out <- frags[i]
		}
	}
}

// parseIIPData extracts the data value from an IIP data token.
func parseIIPData(dataTok *types.Token) (any, *types.ParseError) {
	//nolint:exhaustive // default handles unexpected token types
	switch dataTok.Type {
	case types.TokQuoted:
		v := dataTok.Value
		if len(v) >= 2 {
			v = v[1 : len(v)-1]
		}

		return v, nil
	case types.TokInt:
		n, err := strconv.Atoi(dataTok.Value)
		if err != nil {
			return nil, &types.ParseError{
				Span: dataTok.Span,
				Err:  fmt.Errorf("invalid integer %q: %w", dataTok.Value, err),
			}
		}

		return n, nil
	default:
		return nil, &types.ParseError{
			Span: dataTok.Span,
			Err:  fmt.Errorf("expected quoted string or integer for IIP data, got %s %q", dataTok.Type, dataTok.Value),
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
func parseIIPStatement(stmt types.Statement) ([]types.Fragment, *types.ParseError) {
	cur := newTokenCursor(stmt.Tokens)

	dataTok := cur.consume()

	data, err := parseIIPData(&dataTok)
	if err != nil {
		return nil, err
	}

	if _, err := cur.expect(types.TokArrow); err != nil {
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

	var frags []types.Fragment

	if tgtComp != "" {
		frags = append(frags, types.Fragment{
			Kind:    types.FragmentProcess,
			Process: &types.ProcessDef{Name: tgtProcTok.Value, Component: tgtComp},
		})
	}

	frags = append(frags, types.Fragment{
		Kind: types.FragmentIIP,
		IIP: &types.IIPDef{
			Data: data,
			Tgt:  types.Endpoint{Process: tgtProcTok.Value, Port: tgtPortTok.Value, Index: tgtIndex},
		},
	})

	return frags, nil
}
