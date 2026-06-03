# dsl — FBP DSL Parser

Package `dsl` parses [Flow-Based Programming](http://en.wikipedia.org/wiki/Flow-based_programming) (FBP) definition files into a serializable `Definition` and can build a runnable `*goflow.Graph` from it.

## Overview

The parser understands the FBP network definition language:

- **Process declarations** — name a node and assign it a component.
- **Connections** — link an output port to an input port.
- **IIPs** — send an initial value to a port before the network starts.
- **Exports** — expose a process port as a graph-level port.

The package exposes a high-level API that hides the internal lexer / parser pipeline, as well as lower-level helpers for inspecting or caching parsed graphs.

## Installation

The package is part of `github.com/trustmaster/goflow`. Import it as:

```go
import "github.com/trustmaster/goflow/dsl"
```

## FBP Syntax

A minimal `.fbp` file looks like this:

```fbp
# Declare two processes
Sender(test/sender) OUT -> IN Receiver(test/receiver)

# Export the receiver's output as a graph port
OUTPORT=Receiver.OUT:OUT
```

### Supported constructs

| Construct | Example | Description |
|---|---|---|
| Process declaration | `ProcName(pkg/component)` | Creates a process named `ProcName` using the registered component `pkg/component`. |
| Connection | `Src OUT -> IN Tgt` | Connects `Src.OUT` to `Tgt.IN`. |
| Array port | `Src OUT -> IN[0] Tgt` | Connects to index `0` of an array port. |
| IIP | `"hello" -> IN Tgt` | Sends the string `"hello"` to `Tgt.IN` on startup. |
| In-port export | `INPORT=Tgt.IN:INPUT` | Exposes `Tgt.IN` as the graph-level input port `INPUT`. |
| Out-port export | `OUTPORT=Src.OUT:OUTPUT` | Exposes `Src.OUT` as the graph-level output port `OUTPUT`. |
| Comment | `# this is a comment` | Everything after `#` until end of line is ignored. |

## Usage

### Parse an FBP file and run it

```go
package main

import (
    "fmt"

    "github.com/trustmaster/goflow"
    "github.com/trustmaster/goflow/dsl"
)

func main() {
    // 1. Create a factory and register your components.
    f := goflow.NewFactory()
    f.Register("test/sender", func() (interface{}, error) { return new(Sender), nil })
    f.Register("test/receiver", func() (interface{}, error) { return new(Receiver), nil })

    // 2. Load the FBP file and build a runnable graph.
    g, err := dsl.LoadFile("graph.fbp", f)
    if err != nil {
        panic(err)
    }

    // 3. Bind to exported ports and run.
    out := make(chan int, 1)
    if err := g.SetOutPort("OUT", out); err != nil {
        panic(err)
    }

    wait := goflow.Run(g)
    fmt.Println(<-out)
    <-wait
}
```

### Parse source bytes directly

```go
src := []byte(`
Sender(test/sender) OUT -> IN Receiver(test/receiver)
OUTPORT=Receiver.OUT:OUT
`)

g, err := dsl.Parse(src, f)
```

### Inspect a parsed Definition

If you only need the structural description (processes, connections, exports, etc.) without building a graph:

```go
src := []byte(`Sender(test/sender) OUT -> IN Receiver(test/receiver)`)

def, err := dsl.ParseDefinition(src)
if err != nil {
    panic(err)
}

fmt.Printf("Processes: %d\n", len(def.Processes))
fmt.Printf("Connections: %d\n", len(def.Connections))
fmt.Printf("Sender component: %s\n", def.Processes["Sender"].Component)
```

### Cache a Definition as JSON

`Definition` is serializable, so you can parse once and reuse later without re-running the parser:

```go
// Parse once.
def, err := dsl.ParseDefinition(src)
if err != nil {
    panic(err)
}

// Marshal to JSON for storage or transport.
data, err := json.Marshal(def)
if err != nil {
    panic(err)
}

// Later, unmarshal and build.
cached, err := dsl.UnmarshalDefinition(data)
if err != nil {
    panic(err)
}

g, err := dsl.Build(cached, f)
```

## Error handling

The parser returns three kinds of errors:

- **`dsl.LexError`** — invalid tokens or unexpected characters.
- **`dsl.ParseError`** — syntactically invalid statements.
- **`dsl.BuildError`** — validation failures while constructing the graph (missing processes, bad port names, etc.).

All three implement the standard `error` interface.

## Types

See [definition.go](definition.go) for the core data structures:

- `Definition` — the top-level graph description.
- `ProcessDef` — a named process and its component.
- `ConnectionDef` — a directed edge between two ports.
- `IIPDef` — an initial information packet.
- `ExportDef` — a graph-level port export.

## Internal architecture

The parser is itself implemented as a GoFlow network: a lexer tokenizes the input, a strip-trivia stage removes whitespace and comments, a segmenter groups tokens into statements, and finally a parser emits `Fragment`s that are collected into a `Definition`. This design is an example of dog-fooding — the FBP parser is built with the FBP framework it serves.

The graph structure is described in `.fbp` files that mirror the runtime implementation:

- **[`dsl.fbp`](dsl.fbp)** — the top-level pipeline: `Lexer` → `StripTrivia` → `SegmentStatements` → `Parser`.
- **[`lexer.fbp`](lexer.fbp)** — the lexer subgraph: `StartCursor` → `Dispatch` → scanner components → `Advance`.
- **[`parser.fbp`](parser.fbp)** — the parser subgraph: `RouteStatements` → `ParseExport` / `ParseIIP` / `ParseConnection` → `CollectDefinition`.

For most users the internal pipeline is an implementation detail; the public API (`Parse`, `LoadFile`, `ParseDefinition`, etc.) handles wiring and execution automatically.
