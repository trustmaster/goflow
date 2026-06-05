package parse_test

import (
	"sync"
	"testing"
	"time"

	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/parse"
	"github.com/trustmaster/goflow/dsl/types"
)

// ---------------------------------------------------------------------------
// Test components (registered via factory for Build tests)
// ---------------------------------------------------------------------------

// testReceiver passes int values from In to Out.
type testReceiver struct {
	In  <-chan int
	Out chan<- int
}

func (c *testReceiver) Process() {
	for v := range c.In {
		c.Out <- v
	}
}

// testArrayReceiver reads from an array input port and forwards values to Out.
type testArrayReceiver struct {
	In  [](<-chan int)
	Out chan<- int
}

func (c *testArrayReceiver) Process() {
	var wg sync.WaitGroup

	for _, ch := range c.In {
		wg.Add(1)

		go func(ch <-chan int) {
			defer wg.Done()

			for v := range ch {
				c.Out <- v
			}
		}(ch)
	}

	wg.Wait()
}

// testSender sends a fixed value on Out and exits.
type testSender struct {
	Out chan<- int
}

func (c *testSender) Process() {
	c.Out <- 42
}

// registerTestComponents registers minimal components needed for build tests.
func registerTestComponents(f *goflow.Factory) error {
	comps := []struct {
		name string
		ctor func() (interface{}, error)
	}{
		{"test/receiver", func() (interface{}, error) { return new(testReceiver), nil }},
		{"test/arrayReceiver", func() (interface{}, error) { return new(testArrayReceiver), nil }},
		{"test/sender", func() (interface{}, error) { return new(testSender), nil }},
	}

	for _, c := range comps {
		if err := f.Register(c.name, c.ctor); err != nil {
			return err
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestBuild_EmptyDefinition(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	def := types.Definition{Processes: make(map[string]types.ProcessDef)}
	g, err := parse.Build(&def, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g == nil {
		t.Fatal("expected non-nil graph")
	}
}

func TestBuild_SingleProcess(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	def := types.Definition{
		Processes: map[string]types.ProcessDef{
			"R": {Name: "R", Component: "test/receiver"},
		},
	}

	g, err := parse.Build(&def, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g == nil {
		t.Fatal("expected non-nil graph")
	}
}

func TestBuild_Connection(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	def := types.Definition{
		Processes: map[string]types.ProcessDef{
			"S": {Name: "S", Component: "test/sender"},
			"R": {Name: "R", Component: "test/receiver"},
		},
		Connections: []types.ConnectionDef{
			{
				Src: types.Endpoint{Process: "S", Port: "Out"},
				Tgt: types.Endpoint{Process: "R", Port: "In"},
			},
		},
		Exports: []types.ExportDef{
			{Kind: types.ExportOutPort, Public: "OUT", Proc: "R", Port: "Out"},
		},
	}

	g, err := parse.Build(&def, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if g == nil {
		t.Fatal("expected non-nil graph")
	}

	// Run the graph and verify data flows.
	out := make(chan int, 1)

	if err := g.SetOutPort("OUT", out); err != nil {
		t.Fatalf("SetOutPort: %v", err)
	}

	go g.Process()

	select {
	case v := <-out:
		if v != 42 {
			t.Errorf("expected 42, got %d", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for output")
	}
}

func TestBuild_IIP(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	def := types.Definition{
		Processes: map[string]types.ProcessDef{
			"R": {Name: "R", Component: "test/receiver"},
		},
		IIPs: []types.IIPDef{
			{Data: 99, Tgt: types.Endpoint{Process: "R", Port: "In"}},
		},
		Exports: []types.ExportDef{
			{Kind: types.ExportOutPort, Public: "OUT", Proc: "R", Port: "Out"},
		},
	}

	g, err := parse.Build(&def, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Run the graph and verify the IIP value flows through.
	out := make(chan int, 1)

	if err := g.SetOutPort("OUT", out); err != nil {
		t.Fatalf("SetOutPort: %v", err)
	}

	go g.Process()

	select {
	case v := <-out:
		if v != 99 {
			t.Errorf("expected 99, got %d", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for IIP output")
	}
}

func TestBuild_ArrayPort(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	idx := 0

	def := types.Definition{
		Processes: map[string]types.ProcessDef{
			"S": {Name: "S", Component: "test/sender"},
			"R": {Name: "R", Component: "test/arrayReceiver"},
		},
		Connections: []types.ConnectionDef{
			{
				Src: types.Endpoint{Process: "S", Port: "Out"},
				Tgt: types.Endpoint{Process: "R", Port: "In", Index: &idx},
			},
		},
		Exports: []types.ExportDef{
			{Kind: types.ExportOutPort, Public: "OUT", Proc: "R", Port: "Out"},
		},
	}

	g, err := parse.Build(&def, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Run the graph and verify data flows through the array port.
	out := make(chan int, 1)

	if err := g.SetOutPort("OUT", out); err != nil {
		t.Fatalf("SetOutPort: %v", err)
	}

	go g.Process()

	select {
	case v := <-out:
		if v != 42 {
			t.Errorf("expected 42, got %d", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for array port output")
	}
}

func TestBuild_InPortExport(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	def := types.Definition{
		Processes: map[string]types.ProcessDef{
			"R": {Name: "R", Component: "test/receiver"},
		},
		Exports: []types.ExportDef{
			{Kind: types.ExportInPort, Public: "IN", Proc: "R", Port: "In"},
			{Kind: types.ExportOutPort, Public: "OUT", Proc: "R", Port: "Out"},
		},
	}

	g, err := parse.Build(&def, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Run the graph, send data via inport, verify it arrives on outport.
	in := make(chan int, 1)
	out := make(chan int, 1)

	if err := g.SetInPort("IN", in); err != nil {
		t.Fatalf("SetInPort: %v", err)
	}

	if err := g.SetOutPort("OUT", out); err != nil {
		t.Fatalf("SetOutPort: %v", err)
	}

	go g.Process()

	in <- 77
	close(in)

	select {
	case v := <-out:
		if v != 77 {
			t.Errorf("expected 77, got %d", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for inport output")
	}
}

func TestBuild_ErrorUnknownComponent(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	def := types.Definition{
		Processes: map[string]types.ProcessDef{
			"X": {Name: "X", Component: "test/doesNotExist"},
		},
	}

	_, err := parse.Build(&def, f)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, ok := err.(*types.BuildError); !ok {
		t.Fatalf("expected *BuildError, got %T: %v", err, err)
	}
}

func TestBuild_ErrorUndeclaredProcessInConnection(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	def := types.Definition{
		Processes: map[string]types.ProcessDef{
			"S": {Name: "S", Component: "test/sender"},
		},
		Connections: []types.ConnectionDef{
			{
				Src: types.Endpoint{Process: "S", Port: "Out"},
				Tgt: types.Endpoint{Process: "Missing", Port: "In"},
			},
		},
	}

	_, err := parse.Build(&def, f)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, ok := err.(*types.BuildError); !ok {
		t.Fatalf("expected *BuildError, got %T: %v", err, err)
	}
}

func TestBuild_ErrorUndeclaredProcessInIIP(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	def := types.Definition{
		Processes: map[string]types.ProcessDef{},
		IIPs: []types.IIPDef{
			{Data: 1, Tgt: types.Endpoint{Process: "Missing", Port: "In"}},
		},
	}

	_, err := parse.Build(&def, f)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, ok := err.(*types.BuildError); !ok {
		t.Fatalf("expected *BuildError, got %T: %v", err, err)
	}
}

func TestBuild_ErrorUnknownExportKind(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	def := types.Definition{
		Processes: map[string]types.ProcessDef{
			"R": {Name: "R", Component: "test/receiver"},
		},
		Exports: []types.ExportDef{
			{Kind: "invalid", Public: "X", Proc: "R", Port: "In"},
		},
	}

	_, err := parse.Build(&def, f)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, ok := err.(*types.BuildError); !ok {
		t.Fatalf("expected *BuildError, got %T: %v", err, err)
	}
}

func TestBuild_ErrorExportUnknownProcess(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	def := types.Definition{
		Processes: map[string]types.ProcessDef{},
		Exports: []types.ExportDef{
			{Kind: types.ExportInPort, Public: "X", Proc: "Missing", Port: "In"},
		},
	}

	_, err := parse.Build(&def, f)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, ok := err.(*types.BuildError); !ok {
		t.Fatalf("expected *BuildError, got %T: %v", err, err)
	}
}

func TestBuild_ErrorInvalidConnection(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	// Try to connect to a port that does not exist on the component.
	def := types.Definition{
		Processes: map[string]types.ProcessDef{
			"S": {Name: "S", Component: "test/sender"},
			"R": {Name: "R", Component: "test/receiver"},
		},
		Connections: []types.ConnectionDef{
			{
				Src: types.Endpoint{Process: "S", Port: "Out"},
				Tgt: types.Endpoint{Process: "R", Port: "NonExistent"},
			},
		},
	}

	_, err := parse.Build(&def, f)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, ok := err.(*types.BuildError); !ok {
		t.Fatalf("expected *BuildError, got %T: %v", err, err)
	}
}

func TestBuild_MultipleProcessesAndConnections(t *testing.T) {
	f := goflow.NewFactory()
	_ = registerTestComponents(f)

	// Chain: S -> R1 -> R2, with export on R2.Out
	def := types.Definition{
		Processes: map[string]types.ProcessDef{
			"S":  {Name: "S", Component: "test/sender"},
			"R1": {Name: "R1", Component: "test/receiver"},
			"R2": {Name: "R2", Component: "test/receiver"},
		},
		Connections: []types.ConnectionDef{
			{Src: types.Endpoint{Process: "S", Port: "Out"}, Tgt: types.Endpoint{Process: "R1", Port: "In"}},
			{Src: types.Endpoint{Process: "R1", Port: "Out"}, Tgt: types.Endpoint{Process: "R2", Port: "In"}},
		},
		Exports: []types.ExportDef{
			{Kind: types.ExportOutPort, Public: "OUT", Proc: "R2", Port: "Out"},
		},
	}

	g, err := parse.Build(&def, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := make(chan int, 1)

	if err := g.SetOutPort("OUT", out); err != nil {
		t.Fatalf("SetOutPort: %v", err)
	}

	go g.Process()

	select {
	case v := <-out:
		if v != 42 {
			t.Errorf("expected 42, got %d", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for multi-process output")
	}
}
