package dsl_test

import (
	"encoding/json"
	"fmt"

	"github.com/trustmaster/goflow"
	"github.com/trustmaster/goflow/dsl"
)

// ExampleLoadFile demonstrates loading an .fbp file, building a runnable
// graph, and executing it.
func ExampleLoadFile() {
	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		panic(err)
	}

	g, err := dsl.LoadFile("testdata/hello.fbp", f)
	if err != nil {
		panic(err)
	}

	out := make(chan int, 1)
	if err := g.SetOutPort("OUT", out); err != nil {
		panic(err)
	}

	wait := goflow.Run(g)
	fmt.Println(<-out)
	<-wait

	// Output: 42
}

// ExampleParseDefinition demonstrates parsing FBP source bytes into a
// serializable Definition and inspecting its structure.
func ExampleParseDefinition() {
	src := []byte(`Sender(test/sender) OUT -> IN Receiver(test/receiver)`)

	def, err := dsl.ParseDefinition(src)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Processes: %d\n", len(def.Processes))
	fmt.Printf("Connections: %d\n", len(def.Connections))
	fmt.Printf("Sender component: %s\n", def.Processes["Sender"].Component)

	// Output: Processes: 2
	// Connections: 1
	// Sender component: test/sender
}

// ExampleDefinition_marshalJSON demonstrates serializing a parsed Definition
// to JSON. Because JSON object key order is non-deterministic, this example
// does not include a deterministic Output comment; it simply shows the API.
func ExampleDefinition_marshalJSON() {
	src := []byte(`Sender(test/sender) OUT -> IN Receiver(test/receiver)`)

	def, err := dsl.ParseDefinition(src)
	if err != nil {
		panic(err)
	}

	data, err := json.MarshalIndent(def, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
}

// ExampleUnmarshalDefinition demonstrates the cold-start caching workflow:
// parse once, marshal the Definition to JSON, then unmarshal and build later.
func ExampleUnmarshalDefinition() {
	// Step 1: Parse FBP source into a Definition.
	src := []byte(`Sender(test/sender) OUT -> IN Receiver(test/receiver)
OUTPORT=Receiver.OUT:OUT
`)
	def, err := dsl.ParseDefinition(src)
	if err != nil {
		panic(err)
	}

	// Step 2: Marshal to JSON for caching or transport.
	data, err := json.Marshal(def)
	if err != nil {
		panic(err)
	}

	// Step 3: Later, possibly in another process, unmarshal and build.
	cached, err := dsl.UnmarshalDefinition(data)
	if err != nil {
		panic(err)
	}

	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		panic(err)
	}

	g, err := dsl.Build(cached, f)
	if err != nil {
		panic(err)
	}

	out := make(chan int, 1)
	if err := g.SetOutPort("OUT", out); err != nil {
		panic(err)
	}

	wait := goflow.Run(g)
	fmt.Println(<-out)
	<-wait

	// Output: 42
}

// ExampleLoadDefinitionFile_complex demonstrates parsing a multi-process graph
// that uses connections, array ports, IIPs, and exports.
func ExampleLoadDefinitionFile_complex() {
	def, err := dsl.LoadDefinitionFile("testdata/complex.fbp")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Processes: %d\n", len(def.Processes))
	fmt.Printf("Connections: %d\n", len(def.Connections))
	fmt.Printf("IIPs: %d\n", len(def.IIPs))
	fmt.Printf("Exports: %d\n", len(def.Exports))
	fmt.Printf("Array port index: %d\n", *def.Connections[1].Tgt.Index)
	fmt.Printf("IIP data: %v\n", def.IIPs[0].Data)
	fmt.Printf("INPORT public name: %s\n", def.Exports[0].Public)
	fmt.Printf("OUTPORT public name: %s\n", def.Exports[1].Public)

	// Output: Processes: 5
	// Connections: 3
	// IIPs: 1
	// Exports: 2
	// Array port index: 1
	// IIP data: hello
	// INPORT public name: INPUT
	// OUTPORT public name: OUTPUT
}

// ExampleLoadFile_inport demonstrates a three-stage pipeline with both an
// exported input port and an exported output port.
func ExampleLoadFile_inport() {
	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		panic(err)
	}

	g, err := dsl.LoadFile("testdata/inport_chain.fbp", f)
	if err != nil {
		panic(err)
	}

	in := make(chan int, 1)
	out := make(chan int, 1)
	if err := g.SetInPort("INPUT", in); err != nil {
		panic(err)
	}
	if err := g.SetOutPort("OUTPUT", out); err != nil {
		panic(err)
	}

	wait := goflow.Run(g)
	in <- 77
	close(in)
	fmt.Println(<-out)
	<-wait

	// Output: 77
}

// ExampleLoadFile_iip demonstrates a pipeline seeded by an initial
// information packet sent before the network starts.
func ExampleLoadFile_iip() {
	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		panic(err)
	}

	g, err := dsl.LoadFile("testdata/iip_chain.fbp", f)
	if err != nil {
		panic(err)
	}

	out := make(chan int, 1)
	if err := g.SetOutPort("OUTPUT", out); err != nil {
		panic(err)
	}

	wait := goflow.Run(g)
	fmt.Println(<-out)
	<-wait

	// Output: 99
}

// ExampleLoadFile_arrayPort demonstrates multiple senders feeding distinct
// indices of an array port.
func ExampleLoadFile_arrayPort() {
	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		panic(err)
	}

	g, err := dsl.LoadFile("testdata/arrayport_multi.fbp", f)
	if err != nil {
		panic(err)
	}

	// Buffer of 2 because both senders emit a value.
	out := make(chan int, 2)
	if err := g.SetOutPort("OUT", out); err != nil {
		panic(err)
	}

	wait := goflow.Run(g)
	fmt.Println(<-out)
	<-wait

	// Output: 42
}
