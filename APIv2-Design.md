# API v2 Design Proposal

See also: [`APIv2-Implementation.md`](APIv2-Implementation.md)

## Summary

This document proposes a v2 API for GoFlow that combines:

- **package rename from `goflow` to `flow`** while keeping the Go module path as `github.com/trustmaster/goflow`

- **declarative generic port declarations** on components
- **optional struct-based metadata** for port declarations when advanced semantics are needed
- **reference-based wiring by default** for native Go graph definitions
- **name-based wiring via `...ByName` methods** for runtime/file-loaded graphs

The main design goal is to keep the library declarative and pleasant in native Go code while still preserving the original value of GoFlow: graphs can also be described externally in JSON, YAML, or FBP-style DSLs.

The key shift is this:

- in v2, user code imports the module from `github.com/trustmaster/goflow` and refers to the package as `flow`

- in **Go-native code**, ports are wired by **references**
- in **runtime-loaded graphs**, ports are wired by **names**
- both modes share the same internal port metadata and runtime behavior

## Goals

- Replace reflection-based field discovery with explicit port declarations.
- Keep component definitions declarative and close to plain Go structs.
- Make **reference-based wiring** the primary native Go API.
- Preserve **string-based graph definitions** for loaders and external tools.
- Avoid code generation as a requirement.
- Preserve support for:
  - scalar ports
  - map ports like `In[key]`
  - slice ports like `In[2]`
  - fan-in / fan-out
  - graph inports / outports
  - IIPs
  - subgraphs
- Make channel ownership and close policy explicit.

## Non-goals

- Making file-based graph definitions compile-time type-safe.
- Eliminating all runtime validation at `string`/`any` boundaries.
- Preserving v1 method names unchanged.
- Preserving the v1 package name.
- Hiding all complexity of map/slice addressable ports behind raw field pointers alone.

---

## 1. Core model

### 1.0 Package identity in v2

The v2 API intentionally renames the package itself to:

```go
package flow
```

while keeping the project name and Go module path unchanged:

```go
module github.com/trustmaster/goflow
```

So native code will look like:

```go
import "github.com/trustmaster/goflow"

func example() {
    var _ flow.Graph
}
```

This is a breaking change, but it makes the API cleaner and aligns better with how users talk about the library in code. This also goes back to the root of how the package was called in the old v0 branch.


### 1.1 Components remain plain process types

The `Component` contract stays simple:

```go
type Component interface {
    Process()
}
```

Ports are still plain Go channel fields on structs.

### 1.2 Components declare ports explicitly

Wireable components implement:

```go
type PortAware interface {
    Ports() PortSchema
}
```

Where:

```go
type PortSchema map[string]PortDecl

type PortOpts struct {
    Description string
    Optional    bool
    Addressable bool
    Close       ClosePolicy
}
```

The key is the public port name. The value is a `PortDecl` returned by generic helper functions such as `In`, `OutMap`, or `InSlice`.

For the common case, declarations stay brief:

```go
"In":  flow.In(&c.In),
"Out": flow.Out(&c.Out),
```

For advanced behavior, the same helpers accept an optional metadata struct:

```go
"In":  flow.In(&c.In, &flow.PortOpts{Description: "main input"}),
"Out": flow.Out(&c.Out, &flow.PortOpts{Close: flow.CloseByGraph}),
```

This preserves concise syntax for typical components while keeping richer semantics explicit and idiomatic through a struct literal.

### 1.3 Port bindings are generic at construction time

Recommended public constructors:

```go
func In[T any](target *<-chan T, opts ...*PortOpts) PortDecl
func Out[T any](target *chan<- T, opts ...*PortOpts) PortDecl

func InMap[T any](target *map[string]<-chan T, opts ...*PortOpts) PortDecl
func OutMap[T any](target *map[string]chan<- T, opts ...*PortOpts) PortDecl

func InSlice[T any](target *[]<-chan T, opts ...*PortOpts) PortDecl
func OutSlice[T any](target *[]chan<- T, opts ...*PortOpts) PortDecl
```

Recommended close-policy enum:

```go
type ClosePolicy int

const (
    CloseDefault ClosePolicy = iota
    CloseByGraph
    CloseByComponent
)
```

The helpers should accept zero or one non-nil `*PortOpts`. Passing no opts is equivalent to using defaults.

### 1.4 Example: primitive component

```go
type Echo struct {
    In  <-chan int
    Out chan<- int
}

func (c *Echo) Process() {
    for v := range c.In {
        c.Out <- v
    }
}

func (c *Echo) Ports() flow.PortSchema {
    return flow.PortSchema{
        "In":  flow.In(&c.In),
        "Out": flow.Out(&c.Out),
    }
}
```

### 1.5 Example: addressable ports

```go
type Router struct {
    In  map[string]<-chan int
    Out map[string]chan<- int
}

func (c *Router) Ports() flow.PortSchema {
    return flow.PortSchema{
        "In": flow.InMap(&c.In, &flow.PortOpts{
            Addressable: true,
        }),
        "Out": flow.OutMap(&c.Out, &flow.PortOpts{
            Addressable: true,
            Close:       flow.CloseByComponent,
        }),
    }
}
```

---

## 2. Wiring model

## 2.1 Native Go wiring is reference-based by default

The default graph-building API in Go code should be reference-based:

```go
n.Connect(&e1.Out, &e2.In)
n.MapInPort(&e1.In, &n.In)
n.MapOutPort(&e2.Out, &n.Out)
n.AddIIP(&e1.In, 42)
```

This is the primary user-facing API for graphs written in Go.

### Why this should be the default

- avoids string typos in native code
- works better with refactoring tools
- matches the fact that native graph definitions are mostly static
- keeps graph construction readable and compact

### Important note

Because Go does not support method overloading or generic methods, these APIs are still runtime-validated methods, not compile-time-typed method signatures. That is acceptable: the main win is **reference-based identity**, not perfect compile-time graph validation.

## 2.2 Runtime/file wiring is name-based and explicit

Name-based APIs remain available, but under explicit `...ByName` methods:

```go
n.ConnectByName("e1", "Out", "e2", "In")
n.MapInPortByName("In", "e1", "In")
n.MapOutPortByName("Out", "e2", "Out")
n.AddIIPByName("e1", "In", 42)
```

These methods are primarily intended for:

- JSON/YAML/FBP DSL loaders
- runtime tooling
- external graph editors
- dynamic graph construction

This keeps the mental model simple:

- **native graphs** use references
- **runtime-loaded graphs** use names

---

## 3. Public API sketch

### 3.1 Port declarations

```go
type PortAware interface {
    Ports() PortSchema
}

type PortSchema map[string]PortDecl

type PortOpts struct {
    Description string
    Optional    bool
    Addressable bool
    Close       ClosePolicy
}
```

### 3.2 Native/reference-based graph building

```go
func (n *Graph) Connect(src, dst any) error
func (n *Graph) MapInPort(inner, outer any) error
func (n *Graph) MapOutPort(inner, outer any) error
func (n *Graph) AddIIP(dst any, data any) error
```

### 3.3 Runtime/name-based graph building

```go
func (n *Graph) ConnectByName(senderName, senderPort, receiverName, receiverPort string) error
func (n *Graph) MapInPortByName(publicName, procName, procPort string) error
func (n *Graph) MapOutPortByName(publicName, procName, procPort string) error
func (n *Graph) AddIIPByName(processName, portName string, data any) error
```

### 3.4 External graph port binding

The same split should apply to binding a graph to the outside world:

```go
func (n *Graph) SetInPort(portRef any, channel any) error
func (n *Graph) SetOutPort(portRef any, channel any) error

func (n *Graph) SetInPortByName(name string, channel any) error
func (n *Graph) SetOutPortByName(name string, channel any) error
```

This keeps the model consistent.

---

## 4. Graph ports and native composite graphs

Reference-based graph mapping implies that graph ports themselves can be represented as typed fields on native composite graphs.

Example:

```go
type DoubleEcho struct {
    flow.Graph

    In  <-chan int
    Out chan<- int

    e1 Echo
    e2 Echo
}

func (g *DoubleEcho) Ports() flow.PortSchema {
    return flow.PortSchema{
        "In":  flow.In(&g.In),
        "Out": flow.Out(&g.Out),
    }
}
```

Then native wiring can be expressed directly:

```go
g.Connect(&g.e1.Out, &g.e2.In)
g.MapInPort(&g.e1.In, &g.In)
g.MapOutPort(&g.e2.Out, &g.Out)
```

### Design consequence

A typed composite graph that embeds `Graph` and declares graph-level ports must register its own `Ports()` schema with the embedded graph runtime during initialization.

The exact helper name is an implementation detail covered in [`APIv2-Implementation.md`](APIv2-Implementation.md).

### Dynamic graphs still work

For dynamically loaded graphs, graph ports remain named exports managed through:

- `MapInPortByName(...)`
- `MapOutPortByName(...)`
- `SetInPortByName(...)`
- `SetOutPortByName(...)`

That distinction is intentional and healthy.

---

## 5. Addressable port references

Raw field references are enough for scalar ports:

```go
n.Connect(&e1.Out, &e2.In)
```

But map/slice ports need a selector, because the field pointer alone does not identify a key or index.

Recommended selector helpers:

```go
n.Connect(flow.KeyOut(&r.Out, "left"), &e.In)
n.Connect(flow.IndexOut(&ir.Out, 2), &e.In)
n.AddIIP(flow.Key(&r.In, "control"), 42)
```

These helpers are not a second full API. They are just selector values needed for addressable ports.

### Design principle

- scalar ports should feel natural and direct
- addressable ports may use small helper values where Go syntax requires it

---

## 6. Runtime loaders and external graph files

The original value of GoFlow’s string-based graph wiring should be preserved.

Graphs defined in JSON, YAML, or FBP DSL still work naturally through the `...ByName` APIs.

Example:

```json
{
  "processes": {
    "e1": { "component": "Echo" },
    "e2": { "component": "Echo" }
  },
  "connections": [
    {
      "src": { "process": "e1", "port": "Out" },
      "tgt": { "process": "e2", "port": "In" }
    }
  ]
}
```

A loader can translate this to:

```go
n.ConnectByName("e1", "Out", "e2", "In")
```

with runtime validation against each component’s declared `PortSchema`.

### This is the intended split

- Go code gets the cleaner reference-based API
- loaders and runtime tooling use `...ByName`

Both resolve against the same declared metadata.

---

## 7. Factory and metadata story

Explicit `Ports()` declarations make runtime metadata cleaner.

A factory can obtain port metadata either by:

- creating a temporary instance and calling `Ports()`
- or using an optional future optimization for static metadata

Metadata comes from the optional `PortOpts` associated with each declared port, for example:

- `Description`
- `Optional`
- `Addressable`
- `Close`

This supports:

- loader validation
- runtime UI / tooling
- component introspection
- subgraph metadata

Port metadata should come from declarations, not from reflection-based field discovery.

---

## 8. Validation model

### 8.1 Native/reference-based APIs

Reference-based APIs validate at runtime that:

- the references belong to ports known to the graph
- port directions are compatible
- payload types match
- selectors are valid for the port shape
- mapping direction is legal

### 8.2 Name-based APIs

Name-based APIs validate at runtime that:

- the process exists
- the declared port exists
- address selectors are valid
- port directions match the operation
- payload types are compatible

### 8.3 Scope of type safety

This design intentionally separates:

- **typed declarations** inside component code
- **runtime validation** at graph assembly boundaries

That is the correct boundary for a system that supports external graph files.

---

## 9. Migration story

### v1 to v2 direction

- `Ports() PortSchema` replaces reflective field discovery as the preferred model.
- `Connect(...)` becomes the reference-based native API.
- old string-style graph-building methods move to `ConnectByName(...)` and related `...ByName` APIs.

### Compatibility strategy

A migration period may keep reflective fallback internally for older components that do not yet implement `Ports()`, but that fallback is transitional and not part of the long-term design.

---

## 10. Pros

### API clarity

- one primary native API
- one explicit runtime API
- less mental overhead than dual equal-status APIs

### Better native ergonomics

- no string port names in Go graph definitions
- better refactorability
- more declarative graph code

### Better internals

- no structural reflection required for port discovery
- shared runtime model for native and loaded graphs
- explicit ownership and metadata

### Better tooling story

- dynamic graphs still supported
- loaders remain first-class
- metadata comes from declarations instead of inference

---

## 11. Tradeoffs and limitations

### Go language limitations

Because Go does not support method overloading or generic methods:

- `Connect(&e1.Out, &e2.In)` cannot be made a generic method
- runtime validation remains necessary at the method boundary

### Addressable ports still need selector helpers

Map and slice ports cannot be fully represented by raw field references alone.

### Typed composite graphs need explicit initialization of graph-level port schema

This is an implementation requirement of the embedded graph runtime.

### File-based graphs remain runtime-checked

That is expected and acceptable.

---

## 12. Recommendation

Adopt the following v2 direction:

1. rename the package to `flow` while keeping the module path `github.com/trustmaster/goflow`
2. require explicit `Ports() PortSchema` declarations for wireable components
3. use concise builder calls by default and optional `PortOpts` structs for advanced metadata and close semantics
4. make **reference-based methods** the primary API for native Go graph definitions
5. rename string-based graph-building methods to explicit `...ByName` variants
6. preserve file-based graphs by routing loaders through `...ByName`
7. support addressable ports with small selector helpers
8. unify both APIs under one internal port/runtime model

This is the cleanest way to get both:

- typed, declarative native Go components and graphs
- dynamic, string-based graph definitions for files and runtimes

without requiring code generation.
