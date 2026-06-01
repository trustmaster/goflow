package dsl

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/trustmaster/goflow"
)

// ---------------------------------------------------------------------------
// End-to-end integration tests
// ---------------------------------------------------------------------------

func TestIntegration_MinimalGraph(t *testing.T) {
	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join("testdata", "minimal.fbp")
	g, err := LoadFile(path, f)
	if err != nil {
		t.Fatalf("LoadFile error: %v", err)
	}

	out := make(chan int, 1)
	if err := g.SetOutPort("OUT", out); err != nil {
		t.Fatalf("SetOutPort error: %v", err)
	}

	wait := goflow.Run(g)

	select {
	case v := <-out:
		if v != 42 {
			t.Fatalf("expected 42, got %d", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for output")
	}

	<-wait
}

func TestIntegration_IIPGraph(t *testing.T) {
	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join("testdata", "iip.fbp")
	g, err := LoadFile(path, f)
	if err != nil {
		t.Fatalf("LoadFile error: %v", err)
	}

	out := make(chan int, 1)
	if err := g.SetOutPort("OUT", out); err != nil {
		t.Fatalf("SetOutPort error: %v", err)
	}

	wait := goflow.Run(g)

	select {
	case v := <-out:
		if v != 99 {
			t.Fatalf("expected 99, got %d", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for IIP output")
	}

	<-wait
}

func TestIntegration_ArrayPortGraph(t *testing.T) {
	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join("testdata", "arrayport.fbp")
	g, err := LoadFile(path, f)
	if err != nil {
		t.Fatalf("LoadFile error: %v", err)
	}

	out := make(chan int, 1)
	if err := g.SetOutPort("OUT", out); err != nil {
		t.Fatalf("SetOutPort error: %v", err)
	}

	wait := goflow.Run(g)

	select {
	case v := <-out:
		if v != 42 {
			t.Fatalf("expected 42, got %d", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for array port output")
	}

	<-wait
}

func TestIntegration_ErrorUnknownComponent(t *testing.T) {
	src := `Foo(unknown/component)`

	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		t.Fatal(err)
	}

	_, err := Parse([]byte(src), f)
	if err == nil {
		t.Fatal("expected error for unknown component, got nil")
	}

	if !errors.Is(err, &BuildError{}) {
		var be *BuildError
		if !errors.As(err, &be) {
			t.Fatalf("expected *BuildError, got %T: %v", err, err)
		}
	}
}

func TestIntegration_ErrorSyntax(t *testing.T) {
	src := `INPORT=Reader.FILE
`

	_, err := ParseDefinition([]byte(src))
	if err == nil {
		t.Fatal("expected error for invalid syntax, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "parse error") {
		t.Fatalf("expected parse error message, got: %s", msg)
	}
	if !strings.Contains(msg, ":1:") {
		t.Fatalf("expected line/column info in error, got: %s", msg)
	}
}

func TestIntegration_ErrorConflictingProcess(t *testing.T) {
	src := `Sender(test/sender) OUT -> IN Receiver(test/receiver)
Sender(test/other) OUT -> IN Receiver2(test/receiver)
`

	_, err := ParseDefinition([]byte(src))
	if err == nil {
		t.Fatal("expected error for conflicting process declaration, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "conflicting component") {
		t.Fatalf("expected 'conflicting component' in error, got: %s", msg)
	}
}

func TestIntegration_ErrorInvalidExportTarget(t *testing.T) {
	src := `INPORT=Missing.IN:IN
`

	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		t.Fatal(err)
	}

	_, err := Parse([]byte(src), f)
	if err == nil {
		t.Fatal("expected error for invalid export target, got nil")
	}

	var be *BuildError
	if !errors.As(err, &be) {
		t.Fatalf("expected *BuildError, got %T: %v", err, err)
	}
}
