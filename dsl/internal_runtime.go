package dsl

import (
	"errors"
	"fmt"
	"sync"

	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/internal/graphbuild"
	"github.com/trustmaster/goflow/dsl/lex"
	"github.com/trustmaster/goflow/dsl/parse"
	"github.com/trustmaster/goflow/dsl/types"
)

//nolint:gochecknoglobals // package-level singletons are intentional and immutable after init.
var (
	internalFactoryOnce   sync.Once
	cachedInternalFactory *goflow.Factory
	cachedInternalErr     error
)

// internalFactory returns the package-private factory used to build the DSL's
// own internal graphs. After initialization it is treated as immutable.
func internalFactory() (*goflow.Factory, error) {
	internalFactoryOnce.Do(func() {
		f := goflow.NewFactory()
		if err := RegisterComponents(f); err != nil {
			cachedInternalErr = fmt.Errorf("register components: %w", err)
			return
		}

		cachedInternalFactory = f
	})

	if cachedInternalErr != nil {
		return nil, cachedInternalErr
	}

	return cachedInternalFactory, nil
}

func newInternalLexerGraph() (*goflow.Graph, error) {
	f, err := internalFactory()
	if err != nil {
		return nil, err
	}

	g, err := lex.NewLexer(f)
	if err != nil {
		return nil, fmt.Errorf("create lexer: %w", err)
	}

	return g, nil
}

func newInternalParserGraph() (*goflow.Graph, error) {
	f, err := internalFactory()
	if err != nil {
		return nil, err
	}

	g, err := parse.NewParser(f)
	if err != nil {
		return nil, fmt.Errorf("create parser: %w", err)
	}

	return g, nil
}

func newInternalPipelineGraph() (*goflow.Graph, error) {
	f, err := internalFactory()
	if err != nil {
		return nil, err
	}

	g, err := graphbuild.Build(&generatedTopLevelDefinition, f)
	if err != nil {
		return nil, fmt.Errorf("create pipeline: %w", err)
	}

	return g, nil
}

func runInternalPipeline(file *types.File) (types.DefinitionResult, error) {
	if file == nil {
		return types.DefinitionResult{}, errors.New("file cannot be nil")
	}

	pipeline, err := newInternalPipelineGraph()
	if err != nil {
		return types.DefinitionResult{}, err
	}

	in := make(chan *types.File, 1)
	out := make(chan types.DefinitionResult, 1)

	if err := pipeline.SetInPort("In", in); err != nil {
		return types.DefinitionResult{}, fmt.Errorf("set pipeline in port: %w", err)
	}

	if err := pipeline.SetOutPort("Out", out); err != nil {
		return types.DefinitionResult{}, fmt.Errorf("set pipeline out port: %w", err)
	}

	wait := goflow.Run(pipeline)

	in <- file

	close(in)

	result := <-out

	<-wait

	return result, nil
}
