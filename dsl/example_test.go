package dsl

import (
	"encoding/json"
	"fmt"

	"github.com/trustmaster/goflow"
)

// ExampleLoadFile demonstrates loading an .fbp file, building a runnable
// graph, and executing it.
func ExampleLoadFile() {
	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		panic(err)
	}

	g, err := LoadFile("testdata/hello.fbp", f)
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

	def, err := ParseDefinition(src)
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

	def, err := ParseDefinition(src)
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
	def, err := ParseDefinition(src)
	if err != nil {
		panic(err)
	}

	// Step 2: Marshal to JSON for caching or transport.
	data, err := json.Marshal(def)
	if err != nil {
		panic(err)
	}

	// Step 3: Later (possibly in another process), unmarshal and build.
	cached, err := UnmarshalDefinition(data)
	if err != nil {
		panic(err)
	}

	f := goflow.NewFactory()
	if err := registerTestComponents(f); err != nil {
		panic(err)
	}

	g, err := Build(cached, f)
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
