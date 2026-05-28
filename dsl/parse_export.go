package dsl

import "fmt"

// ParseExport is a FBP component that parses INPORT/OUTPORT export statements.
type ParseExport struct {
	In  <-chan Statement
	Out chan<- Fragment
}

// Process parses each incoming Statement as an export declaration.
func (p *ParseExport) Process() {
	for stmt := range p.In {
		def, err := parseExportStatement(stmt)
		if err != nil {
			p.Out <- Fragment{Kind: FragmentError, Err: err}
			continue
		}

		p.Out <- Fragment{Kind: FragmentExport, Export: &def}
	}
}

// parseExportStatement parses a single export statement.
//
// Expected token sequence:
//
//	INPORT  = ProcName . PortName : PublicName
//	OUTPORT = ProcName . PortName : PublicName
func parseExportStatement(stmt Statement) (ExportDef, *ParseError) {
	cur := newTokenCursor(stmt.Tokens)

	keyword, err := cur.expectIdent()
	if err != nil {
		return ExportDef{}, &ParseError{Span: stmt.Span, Err: fmt.Errorf("export statement must start with INPORT or OUTPORT")}
	}

	var kind ExportKind
	switch keyword.Value {
	case "INPORT":
		kind = ExportInPort
	case "OUTPORT":
		kind = ExportOutPort
	default:
		return ExportDef{}, &ParseError{Span: keyword.Span, Err: fmt.Errorf("expected INPORT or OUTPORT, got %q", keyword.Value)}
	}

	if _, err := cur.expect(TokEqual); err != nil {
		return ExportDef{}, err
	}

	proc, err := cur.expectIdent()
	if err != nil {
		return ExportDef{}, err
	}

	if _, err := cur.expect(TokDot); err != nil {
		return ExportDef{}, err
	}

	port, err := cur.expectIdent()
	if err != nil {
		return ExportDef{}, err
	}

	if _, err := cur.expect(TokColon); err != nil {
		return ExportDef{}, err
	}

	public, err := cur.expectIdent()
	if err != nil {
		return ExportDef{}, err
	}

	return ExportDef{
		Kind:   kind,
		Public: public.Value,
		Proc:   proc.Value,
		Port:   port.Value,
	}, nil
}
