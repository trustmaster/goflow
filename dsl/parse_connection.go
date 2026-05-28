package dsl

// ParseConnection is a FBP component that parses connection statements.
type ParseConnection struct {
	In  <-chan Statement
	Out chan<- Fragment
}

// Process parses each incoming Statement as a connection declaration.
func (p *ParseConnection) Process() {
	for stmt := range p.In {
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

// parseConnectionStatement parses a single connection statement.
//
// Expected token sequence:
//
//	SrcProc [( SrcComp )] SrcPort [[SrcIdx]] -> TgtPort [[TgtIdx]] TgtProc [( TgtComp )]
//
// Process declaration fragments are emitted only when an inline component is present.
// A ConnectionDef fragment is always emitted.
func parseConnectionStatement(stmt Statement) ([]Fragment, *ParseError) {
	cur := newTokenCursor(stmt.Tokens)

	// Source endpoint: ProcName [(Component)] PortName [[Index]]
	srcProcTok, err := cur.expectIdent()
	if err != nil {
		return nil, err
	}

	srcComp, err := parseComponent(cur)
	if err != nil {
		return nil, err
	}

	srcPortTok, err := cur.expectIdent()
	if err != nil {
		return nil, err
	}

	srcIndex, err := parseArrayIndex(cur)
	if err != nil {
		return nil, err
	}

	// Arrow
	if _, err := cur.expect(TokArrow); err != nil {
		return nil, err
	}

	// Target endpoint: PortName [[Index]] ProcName [(Component)]
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

	if srcComp != "" {
		frags = append(frags, Fragment{
			Kind:    FragmentProcess,
			Process: &ProcessDef{Name: srcProcTok.Value, Component: srcComp},
		})
	}

	if tgtComp != "" {
		frags = append(frags, Fragment{
			Kind:    FragmentProcess,
			Process: &ProcessDef{Name: tgtProcTok.Value, Component: tgtComp},
		})
	}

	frags = append(frags, Fragment{
		Kind: FragmentConnection,
		Connection: &ConnectionDef{
			Src: Endpoint{Process: srcProcTok.Value, Port: srcPortTok.Value, Index: srcIndex},
			Tgt: Endpoint{Process: tgtProcTok.Value, Port: tgtPortTok.Value, Index: tgtIndex},
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

	n := 0

	for _, ch := range tok.Value {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		}
	}

	if _, err := cur.expect(TokRBracket); err != nil {
		return nil, err
	}

	return &n, nil
}
