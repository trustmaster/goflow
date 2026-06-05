package dsl_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl"
	"github.com/trustmaster/goflow/dsl/types"
)

func TestParseDefinition(t *testing.T) {
	src := `Read(test/sender) OUT -> IN Echo(test/receiver)
'hello' -> IN Greeter(test/receiver)
INPORT=Read.IN:INPUT
OUTPORT=Echo.OUT:OUTPUT
Split(test/receiver) OUT[0] -> IN Merge(test/receiver)
`
	def, err := dsl.ParseDefinition([]byte(src))
	if err != nil {
		t.Fatalf("ParseDefinition error: %v", err)
	}

	if len(def.Processes) != 5 {
		t.Fatalf("expected 5 processes, got %d", len(def.Processes))
	}

	wantProcesses := map[string]string{
		"Read":    "test/sender",
		"Echo":    "test/receiver",
		"Greeter": "test/receiver",
		"Split":   "test/receiver",
		"Merge":   "test/receiver",
	}

	for name, comp := range wantProcesses {
		p, ok := def.Processes[name]
		if !ok {
			t.Fatalf("missing process %q", name)
		}
		if p.Component != comp {
			t.Fatalf("process %q component: want %q, got %q", name, comp, p.Component)
		}
	}

	if len(def.Connections) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(def.Connections))
	}

	conn0 := def.Connections[0]
	if conn0.Src.Process != "Read" || conn0.Src.Port != "OUT" || conn0.Tgt.Process != "Echo" || conn0.Tgt.Port != "IN" {
		t.Fatalf("unexpected connection[0]: %+v", conn0)
	}

	conn1 := def.Connections[1]
	if conn1.Src.Process != "Split" || conn1.Src.Port != "OUT" || conn1.Src.Index == nil || *conn1.Src.Index != 0 ||
		conn1.Tgt.Process != "Merge" || conn1.Tgt.Port != "IN" {
		t.Fatalf("unexpected connection[1]: %+v", conn1)
	}

	if len(def.IIPs) != 1 {
		t.Fatalf("expected 1 IIP, got %d", len(def.IIPs))
	}

	iip := def.IIPs[0]
	if iip.Data != "hello" || iip.Tgt.Process != "Greeter" || iip.Tgt.Port != "IN" {
		t.Fatalf("unexpected iip: %+v", iip)
	}

	if len(def.Exports) != 2 {
		t.Fatalf("expected 2 exports, got %d", len(def.Exports))
	}

	exp0 := def.Exports[0]
	if exp0.Kind != types.ExportInPort || exp0.Public != "INPUT" || exp0.Proc != "Read" || exp0.Port != "IN" {
		t.Fatalf("unexpected export[0]: %+v", exp0)
	}

	exp1 := def.Exports[1]
	if exp1.Kind != types.ExportOutPort || exp1.Public != "OUTPUT" || exp1.Proc != "Echo" || exp1.Port != "OUT" {
		t.Fatalf("unexpected export[1]: %+v", exp1)
	}
}

func TestParseDefinition_Error(t *testing.T) {
	src := `INPORT=Reader.FILE
`
	_, err := dsl.ParseDefinition([]byte(src))
	if err == nil {
		t.Fatal("expected error for invalid export syntax")
	}
}

func TestLoadDefinitionFile(t *testing.T) {
	src := `Read(test/sender) OUT -> IN Echo(test/receiver)
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.fbp")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	def, err := dsl.LoadDefinitionFile(path)
	if err != nil {
		t.Fatalf("LoadDefinitionFile error: %v", err)
	}

	if len(def.Processes) != 2 {
		t.Fatalf("expected 2 processes, got %d", len(def.Processes))
	}
}

func TestParse(t *testing.T) {
	src := `Sender(test/sender) OUT -> IN Receiver(test/receiver)
OUTPORT=Receiver.OUT:OUT
`
	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		t.Fatal(err)
	}

	g, err := dsl.Parse([]byte(src), f)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	out := make(chan int)
	if err := g.SetOutPort("OUT", out); err != nil {
		t.Fatalf("SetOutPort error: %v", err)
	}

	wait := goflow.Run(g)
	v := <-out
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
	<-wait
}

func TestLoadFile(t *testing.T) {
	src := `Sender(test/sender) OUT -> IN Receiver(test/receiver)
OUTPORT=Receiver.OUT:OUT
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.fbp")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		t.Fatal(err)
	}

	g, err := dsl.LoadFile(path, f)
	if err != nil {
		t.Fatalf("LoadFile error: %v", err)
	}

	out := make(chan int)
	if err := g.SetOutPort("OUT", out); err != nil {
		t.Fatalf("SetOutPort error: %v", err)
	}

	wait := goflow.Run(g)
	v := <-out
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
	<-wait
}

func TestUnmarshalDefinition(t *testing.T) {
	def := &types.Definition{
		Processes: map[string]types.ProcessDef{
			"Read": {Name: "Read", Component: "test/sender"},
		},
		Connections: []types.ConnectionDef{
			{Src: types.Endpoint{Process: "Read", Port: "OUT"}, Tgt: types.Endpoint{Process: "Echo", Port: "IN"}},
		},
		IIPs: []types.IIPDef{
			{Data: "hello", Tgt: types.Endpoint{Process: "Echo", Port: "IN"}},
		},
		Exports: []types.ExportDef{
			{Kind: types.ExportInPort, Public: "INPUT", Proc: "Read", Port: "IN"},
		},
	}

	data, err := json.Marshal(def)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	got, err := dsl.UnmarshalDefinition(data)
	if err != nil {
		t.Fatalf("UnmarshalDefinition error: %v", err)
	}

	if len(got.Processes) != len(def.Processes) {
		t.Fatalf("process count mismatch: want %d, got %d", len(def.Processes), len(got.Processes))
	}

	if len(got.Connections) != len(def.Connections) {
		t.Fatalf("connection count mismatch")
	}

	if len(got.IIPs) != len(def.IIPs) {
		t.Fatalf("iip count mismatch")
	}

	if len(got.Exports) != len(def.Exports) {
		t.Fatalf("export count mismatch")
	}

	if got.Exports[0].Kind != types.ExportInPort || got.Exports[0].Public != "INPUT" {
		t.Fatalf("unexpected export: %+v", got.Exports[0])
	}
}

func TestDefinitionJSONRoundTrip(t *testing.T) {
	idx := 3
	original := types.Definition{
		Processes: map[string]types.ProcessDef{
			"A": {Name: "A", Component: "comp/a"},
		},
		Connections: []types.ConnectionDef{
			{
				Src: types.Endpoint{Process: "A", Port: "OUT", Index: &idx},
				Tgt: types.Endpoint{Process: "B", Port: "IN"},
			},
		},
		IIPs: []types.IIPDef{
			{Data: 42, Tgt: types.Endpoint{Process: "B", Port: "IN"}},
		},
		Exports: []types.ExportDef{
			{Kind: types.ExportOutPort, Public: "RESULT", Proc: "B", Port: "OUT"},
		},
	}

	data, err := json.Marshal(&original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var roundTripped types.Definition
	if err := json.Unmarshal(data, &roundTripped); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if len(roundTripped.Processes) != 1 {
		t.Fatalf("expected 1 process, got %d", len(roundTripped.Processes))
	}

	if len(roundTripped.Connections) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(roundTripped.Connections))
	}

	if roundTripped.Connections[0].Src.Index == nil || *roundTripped.Connections[0].Src.Index != 3 {
		t.Fatalf("expected array index 3, got %v", roundTripped.Connections[0].Src.Index)
	}

	if len(roundTripped.IIPs) != 1 {
		t.Fatalf("expected 1 IIP, got %d", len(roundTripped.IIPs))
	}

	if roundTripped.IIPs[0].Data != float64(42) {
		// JSON numbers unmarshal as float64 by default.
		t.Fatalf("expected IIP data 42, got %v", roundTripped.IIPs[0].Data)
	}

	if len(roundTripped.Exports) != 1 || roundTripped.Exports[0].Public != "RESULT" {
		t.Fatalf("unexpected export: %+v", roundTripped.Exports)
	}
}
