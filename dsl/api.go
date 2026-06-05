package dsl

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/lex"
	"github.com/trustmaster/goflow/dsl/parse"
	"github.com/trustmaster/goflow/dsl/types"
)

// ParseDefinition parses FBP source bytes into a Definition.
func ParseDefinition(src []byte) (*types.Definition, error) {
	if len(src) == 0 {
		return &types.Definition{
			Processes:   make(map[string]types.ProcessDef),
			Connections: []types.ConnectionDef{},
			IIPs:        []types.IIPDef{},
			Exports:     []types.ExportDef{},
		}, nil
	}

	file := &types.File{Name: "<input>", Data: src}

	return parseFile(file)
}

// LoadDefinitionFile reads an FBP file from disk and parses it into a Definition.
func LoadDefinitionFile(path string) (*types.Definition, error) {
	data, err := os.ReadFile(filepath.Clean(path)) // #nosec G304
	if err != nil {
		return nil, fmt.Errorf("load definition file: %w", err)
	}

	file := &types.File{Name: path, Data: data}

	return parseFile(file)
}

// Parse parses FBP source bytes and builds a runnable *goflow.Graph using the
// provided component factory.
func Parse(src []byte, f *goflow.Factory) (*goflow.Graph, error) {
	def, err := ParseDefinition(src)
	if err != nil {
		return nil, err
	}

	return parse.Build(def, f)
}

// LoadFile reads an FBP file from disk and builds a runnable *goflow.Graph
// using the provided component factory.
func LoadFile(path string, f *goflow.Factory) (*goflow.Graph, error) {
	def, err := LoadDefinitionFile(path)
	if err != nil {
		return nil, err
	}

	return parse.Build(def, f)
}

// UnmarshalDefinition unmarshals a JSON-encoded Definition.
func UnmarshalDefinition(data []byte) (*types.Definition, error) {
	var def types.Definition
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

	lexer, err := lex.NewLexer(f)
	if err != nil {
		return nil, nil, fmt.Errorf("create lexer: %w", err)
	}

	parser, err := parse.NewParser(f)
	if err != nil {
		return nil, nil, fmt.Errorf("create parser: %w", err)
	}

	return lexer, parser, nil
}

func wireLexerInOut(lexer *goflow.Graph, fileCh chan *types.File, tokenCh chan types.Token) error {
	if err := lexer.SetInPort("In", fileCh); err != nil {
		return fmt.Errorf("set lexer in port: %w", err)
	}

	if err := lexer.SetOutPort("Out", tokenCh); err != nil {
		return fmt.Errorf("set lexer out port: %w", err)
	}

	return nil
}

func wireParserInOut(parser *goflow.Graph, stmtCh chan types.Statement, resultCh chan types.DefinitionResult) error {
	if err := parser.SetInPort("In", stmtCh); err != nil {
		return fmt.Errorf("set parser in port: %w", err)
	}

	if err := parser.SetOutPort("Out", resultCh); err != nil {
		return fmt.Errorf("set parser out port: %w", err)
	}

	return nil
}

func runIntermediateStages(tokenCh chan types.Token, stmtCh chan types.Statement) {
	strippedCh := make(chan types.Token)

	strip := &parse.StripTrivia{In: tokenCh, Out: strippedCh}
	stripWait := goflow.Run(strip)

	go func() {
		<-stripWait
		close(strippedCh)
	}()

	segment := &parse.SegmentStatements{In: strippedCh, Out: stmtCh}
	segmentWait := goflow.Run(segment)

	go func() {
		<-segmentWait
		close(stmtCh)
	}()
}

func runPipeline(file *types.File) (types.DefinitionResult, error) {
	if file == nil {
		return types.DefinitionResult{}, errors.New("file cannot be nil")
	}

	lexer, parser, err := createLexerParser()
	if err != nil {
		return types.DefinitionResult{}, err
	}

	fileCh := make(chan *types.File, 1)
	tokenCh := make(chan types.Token)
	stmtCh := make(chan types.Statement)
	resultCh := make(chan types.DefinitionResult)

	if err := wireLexerInOut(lexer, fileCh, tokenCh); err != nil {
		return types.DefinitionResult{}, err
	}

	if err := wireParserInOut(parser, stmtCh, resultCh); err != nil {
		return types.DefinitionResult{}, err
	}

	go func() {
		fileCh <- file

		close(fileCh)
	}()

	lexerWait := goflow.Run(lexer)

	runIntermediateStages(tokenCh, stmtCh)

	parserWait := goflow.Run(parser)

	result := <-resultCh

	<-lexerWait
	<-parserWait

	return result, nil
}

// parseFile runs the lexer -> strip trivia -> segment statements -> parser
// pipeline and returns the collected Definition.
func parseFile(file *types.File) (*types.Definition, error) {
	result, err := runPipeline(file)
	if err != nil {
		return nil, err
	}

	if len(result.Errors) > 0 {
		return nil, errors.Join(result.Errors...)
	}

	return &result.Definition, nil
}

// Build builds a runnable *goflow.Graph from a Definition.
// This is a convenience wrapper around parse.Build.
func Build(def *types.Definition, f *goflow.Factory) (*goflow.Graph, error) {
	return parse.Build(def, f)
}
