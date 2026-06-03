package dsl

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/trustmaster/goflow"
)

// ParseDefinition parses FBP source bytes into a Definition.
func ParseDefinition(src []byte) (*Definition, error) {
	if len(src) == 0 {
		return &Definition{
			Processes:   make(map[string]ProcessDef),
			Connections: []ConnectionDef{},
			IIPs:        []IIPDef{},
			Exports:     []ExportDef{},
		}, nil
	}

	file := &File{Name: "<input>", Data: src}

	return parseFile(file)
}

// LoadDefinitionFile reads an FBP file from disk and parses it into a Definition.
func LoadDefinitionFile(path string) (*Definition, error) {
	data, err := os.ReadFile(filepath.Clean(path)) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("load definition file: %w", err)
	}

	file := &File{Name: path, Data: data}

	return parseFile(file)
}

// Parse parses FBP source bytes and builds a runnable *goflow.Graph using the
// provided component factory.
func Parse(src []byte, f *goflow.Factory) (*goflow.Graph, error) {
	def, err := ParseDefinition(src)
	if err != nil {
		return nil, err
	}

	return Build(def, f)
}

// LoadFile reads an FBP file from disk and builds a runnable *goflow.Graph
// using the provided component factory.
func LoadFile(path string, f *goflow.Factory) (*goflow.Graph, error) {
	def, err := LoadDefinitionFile(path)
	if err != nil {
		return nil, err
	}

	return Build(def, f)
}

// UnmarshalDefinition unmarshals a JSON-encoded Definition.
func UnmarshalDefinition(data []byte) (*Definition, error) {
	var def Definition
	if err := json.Unmarshal(data, &def); err != nil {
		return nil, err
	}

	return &def, nil
}

func createLexerParser() (*goflow.Graph, *goflow.Graph, error) {
	f := goflow.NewFactory()
	if err := RegisterComponents(f); err != nil {
		return nil, nil, fmt.Errorf("register components: %w", err)
	}

	lexer, err := NewLexer(f)
	if err != nil {
		return nil, nil, fmt.Errorf("create lexer: %w", err)
	}

	parser, err := NewParser(f)
	if err != nil {
		return nil, nil, fmt.Errorf("create parser: %w", err)
	}

	return lexer, parser, nil
}

func runPipeline(file *File) (DefinitionResult, error) {
	if file == nil {
		return DefinitionResult{}, errors.New("file cannot be nil")
	}

	lexer, parser, err := createLexerParser()
	if err != nil {
		return DefinitionResult{}, err
	}

	fileCh := make(chan *File, 1)
	tokenCh := make(chan Token)
	strippedCh := make(chan Token)
	stmtCh := make(chan Statement)
	resultCh := make(chan DefinitionResult)

	if err := lexer.SetInPort("In", fileCh); err != nil {
		return DefinitionResult{}, fmt.Errorf("set lexer in port: %w", err)
	}

	if err := lexer.SetOutPort("Out", tokenCh); err != nil {
		return DefinitionResult{}, fmt.Errorf("set lexer out port: %w", err)
	}

	if err := parser.SetInPort("In", stmtCh); err != nil {
		return DefinitionResult{}, fmt.Errorf("set parser in port: %w", err)
	}

	if err := parser.SetOutPort("Out", resultCh); err != nil {
		return DefinitionResult{}, fmt.Errorf("set parser out port: %w", err)
	}

	go func() {
		fileCh <- file

		close(fileCh)
	}()

	lexerWait := goflow.Run(lexer)

	strip := &StripTrivia{In: tokenCh, Out: strippedCh}
	stripWait := goflow.Run(strip)

	go func() {
		<-stripWait
		close(strippedCh)
	}()

	segment := &SegmentStatements{In: strippedCh, Out: stmtCh}
	segmentWait := goflow.Run(segment)

	go func() {
		<-segmentWait
		close(stmtCh)
	}()

	parserWait := goflow.Run(parser)

	result := <-resultCh

	<-lexerWait
	<-parserWait

	return result, nil
}

// parseFile runs the lexer -> strip trivia -> segment statements -> parser
// pipeline and returns the collected Definition.
func parseFile(file *File) (*Definition, error) {
	result, err := runPipeline(file)
	if err != nil {
		return nil, err
	}

	if len(result.Errors) > 0 {
		return nil, errors.Join(result.Errors...)
	}

	return &result.Definition, nil
}
