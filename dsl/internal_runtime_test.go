package dsl

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/trustmaster/goflow/dsl/types"
)

func TestGeneratedTopLevelDefinition(t *testing.T) {
	if len(generatedTopLevelDefinition.Processes) != 4 {
		t.Fatalf("expected 4 top-level processes, got %d", len(generatedTopLevelDefinition.Processes))
	}

	wantComponents := map[string]string{
		"Lexer":             "dsl/Lexer",
		"StripTrivia":       "dsl/StripTrivia",
		"SegmentStatements": "dsl/SegmentStatements",
		"Parser":            "dsl/Parser",
	}

	for name, component := range wantComponents {
		proc, ok := generatedTopLevelDefinition.Processes[name]
		if !ok {
			t.Fatalf("missing generated process %q", name)
		}
		if proc.Component != component {
			t.Fatalf("process %q component: want %q, got %q", name, component, proc.Component)
		}
	}

	if len(generatedTopLevelDefinition.Connections) != 3 {
		t.Fatalf("expected 3 top-level connections, got %d", len(generatedTopLevelDefinition.Connections))
	}

	if len(generatedTopLevelDefinition.Exports) != 2 {
		t.Fatalf("expected 2 top-level exports, got %d", len(generatedTopLevelDefinition.Exports))
	}
}

func TestInternalGraphConstructors(t *testing.T) {
	constructors := []struct {
		name string
		fn   func() error
	}{
		{
			name: "lexer",
			fn: func() error {
				_, err := newInternalLexerGraph()
				return err
			},
		},
		{
			name: "parser",
			fn: func() error {
				_, err := newInternalParserGraph()
				return err
			},
		},
		{
			name: "pipeline",
			fn: func() error {
				_, err := newInternalPipelineGraph()
				return err
			},
		},
	}

	for i := range constructors {
		t.Run(constructors[i].name, func(t *testing.T) {
			if err := constructors[i].fn(); err != nil {
				t.Fatalf("build %s graph: %v", constructors[i].name, err)
			}
		})
	}
}

func TestRunInternalPipelineCompletes(t *testing.T) {
	file := &types.File{
		Name: "test.fbp",
		Data: []byte("Read(test/sender) OUT -> IN Echo(test/receiver)\nOUTPORT=Echo.OUT:OUT\n"),
	}

	type outcome struct {
		result types.DefinitionResult
		err    error
	}

	done := make(chan outcome, 1)
	go func() {
		result, err := runInternalPipeline(file)
		done <- outcome{result: result, err: err}
	}()

	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("runInternalPipeline error: %v", got.err)
		}
		if len(got.result.Errors) != 0 {
			t.Fatalf("unexpected pipeline errors: %v", got.result.Errors)
		}
		if len(got.result.Definition.Processes) != 2 {
			t.Fatalf("expected 2 processes, got %d", len(got.result.Definition.Processes))
		}
		if len(got.result.Definition.Exports) != 1 {
			t.Fatalf("expected 1 export, got %d", len(got.result.Definition.Exports))
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for internal pipeline completion")
	}
}

func TestParseDefinitionConcurrent(t *testing.T) {
	src := []byte("Read(test/sender) OUT -> IN Echo(test/receiver)\nOUTPORT=Echo.OUT:OUT\n")

	const workers = 16

	errs := make(chan error, workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			def, err := ParseDefinition(src)
			if err != nil {
				errs <- err
				return
			}
			if len(def.Processes) != 2 {
				errs <- fmt.Errorf("expected 2 processes, got %d", len(def.Processes))
			}
			if len(def.Exports) != 1 {
				errs <- fmt.Errorf("expected 1 export, got %d", len(def.Exports))
			}
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatal(err)
	}
}
