package parse

import (
	"testing"

	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl/types"
)

// runCollectDefinitionCase wires a fresh CollectDefinition, sends the given fragments,
// closes the input, waits for completion, and returns the accumulated DefinitionResult.
func runCollectDefinitionCase(t *testing.T, fragments []types.Fragment) types.DefinitionResult {
	t.Helper()

	cd := &CollectDefinition{}
	in := make(chan types.Fragment, len(fragments)+1)
	out := make(chan types.DefinitionResult, 1)
	cd.In = in
	cd.Out = out

	wait := goflow.Run(cd)

	for _, f := range fragments {
		in <- f
	}

	close(in)
	<-wait

	return <-out
}

func TestMergeProcess(t *testing.T) {
	t.Run("insert new process", func(t *testing.T) {
		def := &types.Definition{Processes: make(map[string]types.ProcessDef)}

		if err := mergeProcess(def, &types.ProcessDef{Name: "Reader", Component: "dsl/Reader"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, ok := def.Processes["Reader"]
		if !ok {
			t.Fatal("process not stored after insertion")
		}

		if got.Component != "dsl/Reader" {
			t.Errorf("component: want %q, got %q", "dsl/Reader", got.Component)
		}
	})

	t.Run("consistent duplicate is accepted", func(t *testing.T) {
		def := &types.Definition{Processes: make(map[string]types.ProcessDef)}
		proc := &types.ProcessDef{Name: "Reader", Component: "dsl/Reader"}

		_ = mergeProcess(def, proc)

		if err := mergeProcess(def, proc); err != nil {
			t.Fatalf("unexpected error on consistent duplicate: %v", err)
		}

		if len(def.Processes) != 1 {
			t.Errorf("expected 1 process, got %d", len(def.Processes))
		}
	})

	t.Run("conflicting component returns BuildError", func(t *testing.T) {
		def := &types.Definition{Processes: make(map[string]types.ProcessDef)}

		_ = mergeProcess(def, &types.ProcessDef{Name: "P", Component: "comp/A"})

		err := mergeProcess(def, &types.ProcessDef{Name: "P", Component: "comp/B"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if _, ok := err.(*types.BuildError); !ok {
			t.Errorf("expected *types.BuildError, got %T: %v", err, err)
		}
	})
}

func TestCollectDefinition(t *testing.T) {
	t.Run("process fragments are merged", func(t *testing.T) {
		frags := []types.Fragment{
			{Kind: types.FragmentProcess, Process: &types.ProcessDef{Name: "Reader", Component: "dsl/Reader"}},
			{Kind: types.FragmentProcess, Process: &types.ProcessDef{Name: "Parser", Component: "dsl/Parser"}},
		}

		result := runCollectDefinitionCase(t, frags)

		if len(result.Errors) != 0 {
			t.Fatalf("unexpected errors: %v", result.Errors)
		}

		if len(result.Definition.Processes) != 2 {
			t.Fatalf("expected 2 processes, got %d", len(result.Definition.Processes))
		}

		if p, ok := result.Definition.Processes["Reader"]; !ok || p.Component != "dsl/Reader" {
			t.Errorf("Reader: want component %q, got %+v", "dsl/Reader", p)
		}

		if p, ok := result.Definition.Processes["Parser"]; !ok || p.Component != "dsl/Parser" {
			t.Errorf("Parser: want component %q, got %+v", "dsl/Parser", p)
		}
	})

	t.Run("consistent duplicate process", func(t *testing.T) {
		proc := &types.ProcessDef{Name: "Reader", Component: "dsl/Reader"}
		frags := []types.Fragment{
			{Kind: types.FragmentProcess, Process: proc},
			{Kind: types.FragmentProcess, Process: proc},
		}

		result := runCollectDefinitionCase(t, frags)

		if len(result.Errors) != 0 {
			t.Fatalf("unexpected errors: %v", result.Errors)
		}

		if len(result.Definition.Processes) != 1 {
			t.Errorf("expected 1 process, got %d", len(result.Definition.Processes))
		}
	})

	t.Run("conflicting process declarations", func(t *testing.T) {
		frags := []types.Fragment{
			{Kind: types.FragmentProcess, Process: &types.ProcessDef{Name: "P", Component: "comp/A"}},
			{Kind: types.FragmentProcess, Process: &types.ProcessDef{Name: "P", Component: "comp/B"}},
		}

		result := runCollectDefinitionCase(t, frags)

		if len(result.Errors) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(result.Errors), result.Errors)
		}

		if _, ok := result.Errors[0].(*types.BuildError); !ok {
			t.Errorf("expected *types.BuildError, got %T", result.Errors[0])
		}
	})

	t.Run("connection and IIP fragments", func(t *testing.T) {
		frags := []types.Fragment{
			{Kind: types.FragmentConnection, Connection: &types.ConnectionDef{
				Src: types.Endpoint{Process: "Reader", Port: "OUT"},
				Tgt: types.Endpoint{Process: "Parser", Port: "IN"},
			}},
			{Kind: types.FragmentIIP, IIP: &types.IIPDef{
				Data: "hello",
				Tgt:  types.Endpoint{Process: "Parser", Port: "CONFIG"},
			}},
		}

		result := runCollectDefinitionCase(t, frags)

		if len(result.Errors) != 0 {
			t.Fatalf("unexpected errors: %v", result.Errors)
		}

		if len(result.Definition.Connections) != 1 {
			t.Errorf("expected 1 connection, got %d", len(result.Definition.Connections))
		}

		if len(result.Definition.IIPs) != 1 {
			t.Errorf("expected 1 IIP, got %d", len(result.Definition.IIPs))
		}
	})

	t.Run("export fragment", func(t *testing.T) {
		frags := []types.Fragment{
			{Kind: types.FragmentExport, Export: &types.ExportDef{
				Kind:   types.ExportInPort,
				Public: "FILE",
				Proc:   "Reader",
				Port:   "FILE",
			}},
		}

		result := runCollectDefinitionCase(t, frags)

		if len(result.Errors) != 0 {
			t.Fatalf("unexpected errors: %v", result.Errors)
		}

		if len(result.Definition.Exports) != 1 {
			t.Fatalf("expected 1 export, got %d", len(result.Definition.Exports))
		}

		got := result.Definition.Exports[0]

		if got.Kind != types.ExportInPort {
			t.Errorf("export kind: want %q, got %q", types.ExportInPort, got.Kind)
		}

		if got.Public != "FILE" {
			t.Errorf("export public: want %q, got %q", "FILE", got.Public)
		}

		if got.Proc != "Reader" {
			t.Errorf("export proc: want %q, got %q", "Reader", got.Proc)
		}

		if got.Port != "FILE" {
			t.Errorf("export port: want %q, got %q", "FILE", got.Port)
		}
	})

	t.Run("mixed full definition", func(t *testing.T) {
		frags := []types.Fragment{
			{Kind: types.FragmentProcess, Process: &types.ProcessDef{Name: "Reader", Component: "dsl/Reader"}},
			{Kind: types.FragmentConnection, Connection: &types.ConnectionDef{
				Src: types.Endpoint{Process: "Reader", Port: "OUT"},
				Tgt: types.Endpoint{Process: "Parser", Port: "IN"},
			}},
			{Kind: types.FragmentIIP, IIP: &types.IIPDef{
				Data: 42,
				Tgt:  types.Endpoint{Process: "Reader", Port: "CONFIG"},
			}},
			{Kind: types.FragmentExport, Export: &types.ExportDef{
				Kind:   types.ExportInPort,
				Public: "FILE",
				Proc:   "Reader",
				Port:   "FILE",
			}},
		}

		result := runCollectDefinitionCase(t, frags)

		if len(result.Errors) != 0 {
			t.Fatalf("unexpected errors: %v", result.Errors)
		}

		if len(result.Definition.Processes) != 1 {
			t.Errorf("expected 1 process, got %d", len(result.Definition.Processes))
		}

		if len(result.Definition.Connections) != 1 {
			t.Errorf("expected 1 connection, got %d", len(result.Definition.Connections))
		}

		if len(result.Definition.IIPs) != 1 {
			t.Errorf("expected 1 IIP, got %d", len(result.Definition.IIPs))
		}

		if len(result.Definition.Exports) != 1 {
			t.Errorf("expected 1 export, got %d", len(result.Definition.Exports))
		}
	})
}
