# API v2 Implementation Spec

See also: [`APIv2-Design.md`](APIv2-Design.md)

## Summary

This document describes the implementation strategy for the v2 API proposed in [`APIv2-Design.md`](APIv2-Design.md).

In v2, the package clause becomes:

```go
package flow
```

while the module path remains:

```go
module github.com/trustmaster/goflow
```

This document assumes that public types such as `Graph`, `PortSchema`, and `PortDecl` live in package `flow` at that unchanged module path.

The core implementation idea is:

- capture explicit port metadata from `Ports() PortSchema`
- index ports in two ways:
  - **by reference** for native Go wiring
  - **by normalized name** for runtime/file wiring
- resolve both forms into one internal `resolvedPort` model
- run all wiring, mapping, IIP, and external-port binding through the same internal code paths

The main reflection-heavy behavior to eliminate is **struct field discovery and manipulation**. Runtime type metadata may still be captured from generic declarations, but that is not the same as reflective field scanning.

---

## 1. Scope and assumptions

### In scope

- declarative port schema capture from components and typed composite graphs
- reference-based resolution
- `...ByName` resolution
- channel creation / reuse / ownership
- fan-in / fan-out behavior
- graph port mapping
- IIP dispatch
- loader integration
- migration hooks
- package rename handling in docs, examples, and public API shape

### Out of scope

- exact JSON/YAML/FBP DSL syntax
- backward-compatible shims for every removed v1 symbol
- code generation

---

## 2. Runtime data model

## 2.1 `PortDecl`

`PortDecl` is the public schema entry stored in `PortSchema` and returned from helper builders such as `In`, `OutMap`, and `InSlice`.

Recommended shape:

```go
type PortDecl struct {
    binding PortBinding
    opts    PortOpts
}
```

Users normally do not construct `PortDecl` directly. They get one from a builder call like:

```go
flow.In(&c.In)
flow.OutMap(&c.Out, &flow.PortOpts{Addressable: true})
```

## 2.2 `PortOpts`

`PortOpts` carries optional declarative metadata and behavioral policy.

Recommended shape:

```go
type PortOpts struct {
    Description string
    Optional    bool
    Addressable bool
    Close       ClosePolicy
}
```

This keeps the common path concise while still using a plain struct literal for advanced semantics.

## 2.3 `PortBinding`

`PortBinding` is the non-generic runtime binding descriptor captured by generic builders such as `In`, `OutMap`, and `InSlice`.

Recommended shape:

```go
type PortBinding struct {
    dir         portDir
    shape       portShape
    payloadType reflect.Type

    bindInternal func(sel addressSelector, existing channelHandle, bufSize int) (channelHandle, error)
    bindExternal func(sel addressSelector, external any) (channelHandle, error)
    sendIIP      func(sel addressSelector, data any, bufSize int) (channelHandle, error)
}
```

### Notes

- `payloadType` can be captured with `reflect.TypeFor[T]()` or equivalent type-token logic.
- This is acceptable because it is runtime type metadata, not reflective field discovery.
- `PortDecl` carries both runtime binding and optional declarative metadata.
- `PortBinding` carries executable runtime behavior.
- The function fields let the runtime stay non-generic after declaration construction.

## 2.4 `PortSchema`

```go
type PortSchema map[string]PortDecl
```

Port names are normalized on ingestion into the graph runtime.

## 2.5 `resolvedPort`

Both reference-based and name-based lookup should resolve to one internal representation.

```go
type resolvedPort struct {
    scope    portScope
    procName string
    portName string
    decl     PortDecl
    selector addressSelector
}
```

Where `scope` distinguishes at least:

- process port
- graph port

## 2.6 Process entry

```go
type procEntry struct {
    name        string
    instance    any
    portsByName map[string]PortDecl
}
```

## 2.7 Graph state additions

`Graph` should add internal indexes conceptually similar to:

```go
type Graph struct {
    ...

    procs map[string]procEntry

    graphPortsByName map[string]PortDecl

    portsByRef map[any]resolvedPort
    procPortsByName map[string]map[string]PortDecl

    inExports  map[string]resolvedPort
    outExports map[string]resolvedPort
}
```

### Purpose of each index

- `portsByRef`: lookup for native `Connect`, `MapInPort`, `AddIIP`, etc.
- `procPortsByName`: lookup for `ConnectByName` and loader operations
- `graphPortsByName`: typed composite graph-level ports by normalized name
- `inExports` / `outExports`: named public graph ports used by runtime-facing `SetInPortByName` and `SetOutPortByName`

The exact field layout can vary, but the logical indexes should remain.

---

## 3. Schema capture

## 3.1 Capturing component schemas in `Add`

When `Graph.Add(name, c)` is called:

1. validate that `c` is an allowed process type
2. if `c` implements `PortAware`, call `Ports()` exactly once
3. validate and normalize the schema
4. store normalized declarations in `procPortsByName[name]`
5. register reverse lookup entries in `portsByRef`

### Important rule

`Ports()` is treated as **declarative metadata**, not dynamic behavior.

The graph runtime must call it once, cache the result, and ignore later changes.

## 3.2 Capturing graph-owner schema

Reference-based graph port mapping like:

```go
n.MapInPort(&e1.In, &n.In)
```

requires the runtime to know about graph-level ports on the composite owner object.

Because an embedded `Graph` cannot discover the outer struct automatically, v2 needs an explicit owner-binding step.

Recommended API shape:

```go
func (g *Graph) BindOwner(owner PortAware) error
```

This should:

1. call `owner.Ports()` once
2. validate and normalize the graph-level schema
3. populate `graphPortsByName`
4. register graph-port references in `portsByRef`

### Notes

- `BindOwner` is only needed for typed composite graphs.
- Dynamic `*Graph` values loaded from files do not need an owner.
- `BindOwner` should reject repeated incompatible calls.

---

## 4. Schema validation and normalization

All schemas should pass a validation step before they are accepted.

## 4.1 Name normalization

Port names should be normalized using the same effective logic as current `capitalizePortName`, so existing expectations around `in`, `IN`, and `In` remain stable.

## 4.2 Validation rules

Reject schemas that contain:

- empty port names
- duplicate port names after normalization
- `Addressable: true` in `PortOpts` on scalar ports
- more than one non-nil `*PortOpts` passed to a builder
- unsupported close policy for the port direction/shape
- nil or invalid binding targets

## 4.3 Defaulting rules

If opts are omitted, nil, or leave a field at its zero value, apply sensible defaults, for example:

- `Close: CloseDefault` on scalar outputs means graph-closes
- `Close: CloseDefault` on addressable outputs means component-closes
- `Optional: false` means required

These exact defaults should match existing behavior as closely as practical.

---

## 5. Reference model

## 5.1 Scalar port references

Scalar ports can be keyed directly by the captured field pointer passed into the declaration helper:

- `&c.In`
- `&c.Out`
- `&g.In`
- `&g.Out`

These values are comparable and can be used as `map[any]...` keys.

## 5.2 Addressable port references

Map/slice fields need a selector wrapper because the field pointer alone does not identify a key/index.

Recommended helper values:

```go
type keyInRef struct { target any; key string }
type keyOutRef struct { target any; key string }
type indexInRef struct { target any; index int }
type indexOutRef struct { target any; index int }
```

with public constructors such as:

```go
func Key[T any](target *map[string]<-chan T, key string) any
func KeyOut[T any](target *map[string]chan<- T, key string) any
func Index[T any](target *[]<-chan T, index int) any
func IndexOut[T any](target *[]chan<- T, index int) any
```

### Why the helpers are still acceptable

They are not a competing API. They are selector values for the narrow case where raw field pointers are insufficient.

## 5.3 Reference normalization

All reference-based public methods should first normalize their arguments into:

```go
(refKey any, sel addressSelector)
```

then resolve `refKey` in `portsByRef`.

---

## 6. Name-based model

## 6.1 Address parsing

Retain the current `parseAddress` conceptually, but convert the result into:

```go
type addressSelector struct {
    kind  selectorKind
    key   string
    index int
}
```

with normalized port name plus selector.

## 6.2 Name-based lookup

`ConnectByName`, `MapInPortByName`, `MapOutPortByName`, and `AddIIPByName` should:

1. parse and normalize the incoming port string
2. find the process entry
3. look up the normalized port declaration
4. attach the selector
5. continue through the shared internal operation

---

## 7. Public method resolution strategy

## 7.1 `Connect`

```go
func (n *Graph) Connect(src, dst any) error
```

Implementation flow:

1. normalize `src` reference into `(refKey, selector)`
2. normalize `dst` reference into `(refKey, selector)`
3. resolve both through `portsByRef`
4. merge in selectors
5. call shared `connectResolved(src, dst)`

## 7.2 `ConnectByName`

```go
func (n *Graph) ConnectByName(senderName, senderPort, receiverName, receiverPort string) error
```

Implementation flow:

1. resolve sender by process name + port string
2. resolve receiver by process name + port string
3. call shared `connectResolved(src, dst)`

## 7.3 `MapInPort`

```go
func (n *Graph) MapInPort(inner, outer any) error
```

Expected meaning:

- `inner` is a process input port reference
- `outer` is a graph input port reference

Implementation:

1. resolve both refs
2. validate direction and graph/process scope
3. store mapping in a graph-port export structure

## 7.4 `MapInPortByName`

```go
func (n *Graph) MapInPortByName(publicName, procName, procPort string) error
```

Implementation:

1. resolve process input port by name
2. create/update named graph export entry under `publicName`
3. store mapping in `inExports`

## 7.5 `MapOutPort` and `MapOutPortByName`

Mirror the same approach for graph output ports.

## 7.6 `AddIIP`

```go
func (n *Graph) AddIIP(dst any, data any) error
```

Implementation:

1. resolve destination input port reference
2. store pending IIP as a resolved target plus payload
3. on graph start, deliver through shared IIP logic

## 7.7 `AddIIPByName`

Same operation, but destination is resolved from process name + port string.

## 7.8 `SetInPort` / `SetOutPort`

Reference-based external binding should resolve graph-port refs through `portsByRef`, then call shared binding logic.

Name-based variants resolve through `inExports` / `outExports`.

---

## 8. Shared internal operations

The public methods above should funnel into a small set of shared internals.

Recommended internal helpers:

```go
func (n *Graph) connectResolved(src, dst resolvedPort, bufSize int) error
func (n *Graph) mapInResolved(inner, outer resolvedPort) error
func (n *Graph) mapOutResolved(inner, outer resolvedPort) error
func (n *Graph) addIIPResolved(dst resolvedPort, data any) error
func (n *Graph) bindExternalResolved(dst resolvedPort, channel any) error
```

This keeps native and runtime paths behaviorally identical after resolution.

---

## 9. Channel model and ownership

## 9.1 `channelHandle`

A typed runtime wrapper should replace `reflect.Value` channel handling.

```go
type channelHandle interface {
    IncListeners()
    DecListeners() bool
    Close()
}
```

Optionally add internal capabilities such as:

```go
type sendAny interface {
    SendAny(any) error
}
```

## 9.2 Generic implementation

```go
type chanRef[T any] struct {
    ch        chan T
    listeners uint
    ...
}
```

Responsibilities:

- make channels when needed
- reuse channels for fan-in / fan-out
- guard closes
- support typed `SendAny`

## 9.3 Ownership policy

Each binding should track whether the channel is:

- graph-owned
- externally owned
- closed by graph
- closed by component

This replaces the current reflective `closeProcOuts` behavior.

### Recommended compatibility defaults

- graph-created scalar output channels: graph closes on process exit
- addressable output elements: component closes by default
- externally injected channels: do not auto-close unless explicitly owned by graph policy

---

## 10. PortBinding builder behavior

Each generic builder should capture:

- payload type
- pointer to the target field
- port direction
- port shape
- zero or one effective `PortOpts` value
- three closures:
  - `bindInternal`
  - `bindExternal`
  - `sendIIP`

### Example responsibilities

#### `In[T]`

- accept internal channels assignable to `<-chan T`
- accept external channels assignable to `<-chan T` or `chan T`
- create send-capable owned channel when needed for IIP delivery

#### `Out[T]`

- accept internal channels assignable to `chan<- T`
- accept external channels assignable to `chan<- T` or `chan T`

#### `InMap[T]` / `OutMap[T]`

- create maps lazily when needed
- resolve per-key channels
- validate selector kind

#### `InSlice[T]` / `OutSlice[T]`

- grow slices lazily when needed
- resolve per-index channels
- validate selector range rules

---

## 11. Type validation

## 11.1 Connection compatibility

`connectResolved` must validate:

- source is output
- destination is input
- payload types are identical or intentionally assignable by the declared rules
- selectors match port shape

The simplest and safest v2 rule is exact payload-type match.

## 11.2 IIP validation

`sendIIP` should validate the runtime value type against the input payload type and return a contextual error on mismatch.

Recommended error context:

- process / graph port name
- expected payload type
- actual data type

## 11.3 External channel validation

`bindExternal` should accept the directional and bidirectional channel forms that are assignable for the declared direction.

---

## 12. Graph start and IIP dispatch

`Process()` should still:

1. prepare/send IIPs
2. start child components
3. wait for completion
4. close graph-owned outputs safely

### IIP dispatch rules

For each stored IIP:

1. resolve or create the target input channel
2. increment channel listener/ref counts as needed
3. send data through `sendAny`
4. close graph-owned temporary channels only when appropriate

### Important edge case

If a graph input port is bound to an external receive-only channel, the runtime may be unable to inject IIPs into it. The implementation should reject or clearly document this case.

---

## 13. Subgraphs

Subgraphs should also participate through explicit declarations.

### Typed composite subgraphs

- owner calls `BindOwner(...)`
- graph-level ports are available by reference and by normalized name

### Dynamically loaded subgraphs

- graph-level ports exist only as named exports
- `...ByName` methods interact with those exports

### Resolution rule

Name-based operations against subgraphs should resolve through graph export maps, not by reflective recursion into nested component fields.

---

## 14. Loader integration

A runtime/file loader should:

1. create component instances via factory
2. add them to the graph
3. rely on `Add` to capture declared schemas
4. wire them using `ConnectByName`, `MapInPortByName`, `MapOutPortByName`, `AddIIPByName`

This preserves external graph definitions while using explicit runtime metadata instead of reflection-based field discovery.

---

## 15. Factory and metadata extraction

For runtime tooling and UI metadata, the factory can obtain component port info by:

1. creating a temporary instance
2. calling `Ports()` if the instance implements `PortAware`
3. translating the schema into `ComponentInfo`

Mapping rules:

- schema key -> `PortInfo.ID`
- payload type -> `PortInfo.Type`
- `opts.Description` -> `PortInfo.Description`
- `opts.Optional == true` -> `PortInfo.Required = false`
- `opts.Addressable == true` -> `PortInfo.Addressable = true`

This replaces the old commented-out reflection path in `factory.go`.

---

## 16. Migration / compatibility strategy

### Stage 1

- add new schema types and builders
- add `BindOwner`
- add reference-based default methods
- add explicit `...ByName` methods

### Stage 2

- migrate built-in examples and tests
- keep reflective fallback only where necessary for old components

### Stage 3

- deprecate reflective fallback
- remove reflective field discovery from core runtime paths

---

## 17. Testing plan

Add focused tests for:

### Schema capture

- normalized names
- duplicate names
- invalid `PortOpts` combinations
- repeated `BindOwner`

### Reference-based APIs

- scalar `Connect`
- graph-port `MapInPort` / `MapOutPort`
- `AddIIP` by reference
- `SetInPort` / `SetOutPort` by reference

### Name-based APIs

- `ConnectByName`
- `MapInPortByName`
- `MapOutPortByName`
- `AddIIPByName`

### Addressable ports

- map key refs
- slice index refs
- invalid selector kinds
- out-of-range / negative indexes where relevant

### Channel lifecycle

- fan-in close safety
- fan-out reuse
- `CloseByComponent` vs `CloseByGraph` behavior
- IIP plus output close race

### Loader integration

- JSON/YAML/DSL loader uses only `...ByName`
- validation errors are contextual and readable

---

## 18. Suggested file layout

A clean implementation will likely want new focused files such as:

- `ports_decl.go`
  - `PortAware`
  - `PortSchema`
  - `PortDecl`
  - `PortOpts`
  - `ClosePolicy`
- `ports_builders.go`
  - `In`, `Out`, `InMap`, `OutMap`, `InSlice`, `OutSlice`
  - selector helpers like `Key`, `Index`
- `ports_runtime.go`
  - `resolvedPort`
  - `addressSelector`
  - `channelHandle`
  - ref/name resolution logic
- `graph_wiring.go`
  - `Connect`, `ConnectByName`
  - mapping and export logic
- `graph_bind.go`
  - `BindOwner`
  - `SetInPort`, `SetOutPort`
- `graph_iip.go`
  - v2 IIP storage and dispatch

The exact filenames may differ, but separating declaration, resolution, and graph operations will reduce complexity.

---

## 19. Final implementation recommendation

Implement v2 around one runtime truth:

1. every wireable port is explicitly declared in `Ports()` using concise builders with optional `PortOpts`
2. every declared port is indexed by both **reference** and **name**
3. every public wiring API first resolves to `resolvedPort`
4. every actual operation goes through a shared internal path
5. channel lifecycle is explicit and policy-driven instead of reflective

That provides:

- clean native Go ergonomics
- preserved runtime/file graph support
- no required code generation
- no structural reflection in the core wiring path
