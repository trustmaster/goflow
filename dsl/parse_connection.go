package dsl

import (
	"fmt"
	"strconv"
)

// ParseConnection is a FBP component that parses connection statements.
type ParseConnection struct {
	In  <-chan Statement
	Out chan<- Fragment
}

// Process parses each incoming Statement as a connection declaration.
// Standalone process declarations (e.g. "Foo(comp/path)") are also accepted.
func (p *ParseConnection) Process() {
	for stmt := range p.In {
		if proc := tryParseStandaloneProcess(stmt); proc != nil {
			p.Out <- Fragment{Kind: FragmentProcess, Process: proc}
			continue
		}

		frags, err := parseConnectionStatement(stmt)
		if err != nil {
			p.Out <- Fragment{Kind: FragmentError, Err: err}
			continue
		}

		for i := range frags {
			p.Out <- frags[i]
		}
	}
}

// tokenCursorWithError wraps a tokenCursor and tracks parse errors.
type parseEndpointResult struct {
	procName  string
	portName  string
	component string
	index     *int
}

// tryParseStandaloneProcess attempts to parse a statement that consists solely of
// an inline process declaration: ProcName '(' ComponentPath ')'.
// It returns nil if the statement does not match this pattern.
func tryParseStandaloneProcess(stmt Statement) *ProcessDef {
	// Reject statements that contain operators used by connections, IIPs, or exports.
	for i := range stmt.Tokens {
		//nolint:exhaustive // only operator tokens are relevant here
		switch stmt.Tokens[i].Type {
		case TokArrow, TokEqual, TokDot, TokColon:
			return nil
		}
	}

	if len(stmt.Tokens) < 4 {
		return nil
	}

	if stmt.Tokens[0].Type != TokIdent || stmt.Tokens[1].Type != TokLParen {
		return nil
	}

	cur := newTokenCursor(stmt.Tokens)

	procTok, err := cur.expectIdent()
	if err != nil {
		return nil
	}

	comp, err := parseComponent(cur)
	if err != nil || comp == "" {
		return nil
	}

	if !cur.done() {
		return nil
	}

	return &ProcessDef{Name: procTok.Value, Component: comp}
}

// parseSourceEndpoint parses the source side of a connection: ProcName [(Component)] PortName [[Index]].
func parseSourceEndpoint(cur *tokenCursor) (parseEndpointResult, *ParseError) {
	var res parseEndpointResult

	procTok, err := cur.expectIdent()
	if err != nil {
		return res, err
	}

	res.procName = procTok.Value

	res.component, err = parseComponent(cur)
	if err != nil {
		return res, err
	}

	portTok, err := cur.expectIdent()
	if err != nil {
		return res, err
	}

	res.portName = portTok.Value

	res.index, err = parseArrayIndex(cur)
	if err != nil {
		return res, err
	}

	return res, nil
}

// parseTargetEndpoint parses the target side of a connection: PortName [[Index]] ProcName [(Component)].
func parseTargetEndpoint(cur *tokenCursor) (parseEndpointResult, *ParseError) {
	var res parseEndpointResult

	portTok, err := cur.expectIdent()
	if err != nil {
		return res, err
	}

	res.portName = portTok.Value

	res.index, err = parseArrayIndex(cur)
	if err != nil {
		return res, err
	}

	procTok, err := cur.expectIdent()
	if err != nil {
		return res, err
	}

	res.procName = procTok.Value

	res.component, err = parseComponent(cur)
	if err != nil {
		return res, err
	}

	return res, nil
}

// parseConnectionStatement parses a single connection statement.
//
// Expected token sequence:
//
//	SrcProc [(SrcComp)] SrcPort [[SrcIdx]] -> TgtPort [[TgtIdx]] TgtProc [(TgtComp)]
//
// Process declaration fragments are emitted only when an inline component is present.
// A ConnectionDef fragment is always emitted.
func parseConnectionStatement(stmt Statement) ([]Fragment, *ParseError) {
	cur := newTokenCursor(stmt.Tokens)

	src, err := parseSourceEndpoint(cur)
	if err != nil {
		return nil, err
	}

	// Arrow
	if _, err := cur.expect(TokArrow); err != nil {
		return nil, err
	}

	tgt, err := parseTargetEndpoint(cur)
	if err != nil {
		return nil, err
	}

	var frags []Fragment

	if src.component != "" {
		frags = append(frags, Fragment{
			Kind:    FragmentProcess,
			Process: &ProcessDef{Name: src.procName, Component: src.component},
		})
	}

	if tgt.component != "" {
		frags = append(frags, Fragment{
			Kind:    FragmentProcess,
			Process: &ProcessDef{Name: tgt.procName, Component: tgt.component},
		})
	}

	frags = append(frags, Fragment{
		Kind: FragmentConnection,
		Connection: &ConnectionDef{
			Src: Endpoint{Process: src.procName, Port: src.portName, Index: src.index},
			Tgt: Endpoint{Process: tgt.procName, Port: tgt.portName, Index: tgt.index},
		},
	})

	return frags, nil
}

// parseComponent parses an optional inline component declaration: '(' path ')'.
// path is one or more identifiers joined by '/'.
// Returns an empty string and nil if the next token is not '('.
func parseComponent(cur *tokenCursor) (string, *ParseError) {
	if cur.peek().Type != TokLParen {
		return "", nil
	}

	cur.consume() // consume '('

	tok, err := cur.expectIdent()
	if err != nil {
		return "", err
	}

	path := tok.Value

	for cur.peek().Type == TokSlash {
		cur.consume() // consume '/'

		tok, err = cur.expectIdent()
		if err != nil {
			return "", err
		}

		path += "/" + tok.Value
	}

	if _, err := cur.expect(TokRParen); err != nil {
		return "", err
	}

	return path, nil
}

// parseArrayIndex parses an optional array port index: '[' int ']'.
// Returns nil and nil if the next token is not '['.
func parseArrayIndex(cur *tokenCursor) (*int, *ParseError) {
	if cur.peek().Type != TokLBracket {
		return nil, nil
	}

	cur.consume() // consume '['

	tok, err := cur.expect(TokInt)
	if err != nil {
		return nil, err
	}

	n, convErr := strconv.Atoi(tok.Value)
	if convErr != nil {
		return nil, &ParseError{
			Span: tok.Span,
			Err:  fmt.Errorf("invalid integer %q: %w", tok.Value, convErr),
		}
	}

	if _, err := cur.expect(TokRBracket); err != nil {
		return nil, err
	}

	return &n, nil
}
