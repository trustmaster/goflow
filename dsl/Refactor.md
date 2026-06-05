# GoFlow `dsl/` optimization blueprint — phases 1–4

## Scope

This blueprint covers only:

1. **Generate graph definitions from existing `.fbp` files**
2. **Instantiate internal graphs from generated definitions**
3. **Cache internal DSL factory/runtime registration**
4. **Remove manual API-layer glue and run one top-level graph**

## Explicit non-goals

To preserve showcase value, this plan does **not** include:

- merging parser stages into a single parser component
- collapsing the lexer scanner network into a single lexer component
- changing the public `dsl` API
- changing the conceptual architecture away from “parser implemented as GoFlow”

---

# Target end-state

After phase 4:

- `dsl/dsl.fbp`, `dsl/lex/lexer.fbp`, and `dsl/parse/parser.fbp` are the **authoritative topology definitions**
- those `.fbp` files are converted into generated Go definitions at build/dev time
- internal graph construction uses generated definitions + `parse.Build(...)`
- internal component registration is initialized once and reused
- `dsl/api.go` no longer manually stitches lexer/parser/intermediate stages with ad hoc channels/goroutines
- the package still showcases:
  - subgraphs
  - port exports
  - graph-of-components architecture
  - self-hosting via GoFlow

---

# Proposed implementation layout

## Existing files to preserve
- `dsl/dsl.fbp`
- `dsl/lex/lexer.fbp`
- `dsl/parse/parser.fbp`

## New files to add
Suggested shape:

- `dsl/internal_runtime.go`
- `dsl/internal_defs_gen.go`
- `dsl/generate.go`
- `dsl/cmd/dslgen/main.go`

Optional split if preferred:
- `dsl/internal_defs_gen.go`
- `dsl/lex/internal_defs_gen.go`
- `dsl/parse/internal_defs_gen.go`

I recommend a **single generated file in `dsl/`** at first, because it keeps the generator simple and avoids cross-package coupling.

---

# Phase 1 — Generate Go definitions from `.fbp`

## Objective

Make the `.fbp` files the single source of truth for internal topology, while avoiding runtime recursion or runtime parse cost.

## Why this phase first

Right now topology exists twice:

- declaratively in `.fbp`
- imperatively in `dsl/api.go`, `dsl/lex/lexer.go`, `dsl/parse/parser.go`

That duplication is the largest source of code heaviness and drift risk.

## Deliverable

A generator that reads the three internal `.fbp` files and emits Go `types.Definition` values.

## Output contract

The generator should produce named definitions roughly like:

- `generatedTopLevelDefinition`
- `generatedLexerDefinition`
- `generatedParserDefinition`

Each as a `types.Definition` literal.

## Step-by-step

### 1.1 Create a generator entrypoint
Add:

- `dsl/cmd/dslgen/main.go`

Responsibilities:

- read the three `.fbp` files from disk
- convert each into a `types.Definition`
- write a deterministic generated Go file

### 1.2 Choose a generation strategy
Use one of these two strategies:

#### Preferred strategy
Use the public DSL parser itself at generation time:

- parse `dsl/dsl.fbp`
- parse `dsl/lex/lexer.fbp`
- parse `dsl/parse/parser.fbp`

Why this is okay:
- generation is offline, not runtime
- it dogfoods the package
- it validates the internal `.fbp` assets continuously

#### Fallback strategy
If bootstrapping becomes awkward, implement a tiny generator-only loader for these three files.

But I would start with the preferred strategy.

### 1.3 Emit stable Go code
The generator must emit deterministic output.

Important detail:
- `Definition.Processes` is a map, so generation should sort process names before emitting.
- connections/IIPs/exports should preserve declaration order from source.

### 1.4 Add `go:generate`
Add:
- `dsl/generate.go`

Example intent:
- `//go:generate go run ./cmd/dslgen`

### 1.5 Commit the generated file
Generated code should be checked into git.

That keeps:
- normal `go test ./...` working
- consumers independent from generator execution
- diffs reviewable when `.fbp` changes

## Files touched

### New
- `dsl/cmd/dslgen/main.go`
- `dsl/generate.go`
- `dsl/internal_defs_gen.go`

### Read-only inputs
- `dsl/dsl.fbp`
- `dsl/lex/lexer.fbp`
- `dsl/parse/parser.fbp`

## Acceptance criteria

- running `go generate ./dsl` regenerates definitions deterministically
- no handwritten topology data remains necessary to describe the three internal graphs
- generated definitions round-trip to the current graph topology
- generated file is stable across repeated runs when inputs are unchanged

## Validation

Add tests such as:

- `dsl/internal_defs_test.go`
  - assert generated top-level definition contains expected processes/connections/exports
- optional generator snapshot test if you want stronger protection

## Risks

### Risk: generator bootstrapping depends on current runtime topology
Mitigation:
- generation is offline
- keep generator narrow and deterministic
- if needed, temporarily call existing `ParseDefinition`

### Risk: generated code too noisy
Mitigation:
- emit compact but readable literals
- keep only one generated file initially

---

# Phase 2 — Build internal graphs from generated definitions

## Objective

Replace handwritten graph builders with graph construction from generated `types.Definition` values.

## What this removes

Eventually this phase should make these handwritten topology constructors unnecessary:

- `dsl/lex/lexer.go`
- `dsl/parse/parser.go`

And it should prepare removal of top-level manual wiring from `dsl/api.go`.

## Deliverable

Internal constructors that build:

- lexer graph from generated lexer definition
- parser graph from generated parser definition
- top-level DSL pipeline graph from generated top-level definition

using `parse.Build(...)`.

## Important design constraint

Do **not** use runtime DSL parsing to construct the parser internals at runtime.

Instead:
- use **generated definitions**
- then call `parse.Build(def, internalFactory)`

## Step-by-step

### 2.1 Add internal constructor helpers
In `dsl/internal_runtime.go` or `dsl/internal_graphs.go`, add helpers like:

- `newInternalLexerGraph()`
- `newInternalParserGraph()`
- `newInternalPipelineGraph()`

These should each:
- obtain internal factory
- call `parse.Build(...)` on the corresponding generated definition

### 2.2 Refactor `lex.NewLexer`
Current file:
- `dsl/lex/lexer.go`

Current implementation:
- manually adds processes
- manually connects edges
- manually maps ports

Replace implementation with a thin wrapper around generated definition build.

Possible options:

#### Option A
Keep `lex.NewLexer(f)` public and implement it with generated definition build.

#### Option B
Deprecate `lex.NewLexer(f)` internally and route API-layer logic through new internal helpers.

I recommend **Option A** to minimize churn and preserve tests.

### 2.3 Refactor `parse.NewParser`
Current file:
- `dsl/parse/parser.go`

Replace manual process/connection construction with build from generated parser definition.

### 2.4 Add top-level pipeline constructor
Currently `dsl/api.go` manually creates:
- lexer graph
- parser graph
- intermediate stages

Introduce a proper top-level graph constructor using generated `dsl/dsl.fbp` definition.

This is the core bridge to phase 4.

## Files touched

### Modify
- `dsl/lex/lexer.go`
- `dsl/parse/parser.go`

### Add
- `dsl/internal_runtime.go` or `dsl/internal_graphs.go`

### Consume generated file
- `dsl/internal_defs_gen.go`

## Acceptance criteria

- `lex.NewLexer(...)` still returns a working graph
- `parse.NewParser(...)` still returns a working graph
- a new top-level internal DSL graph can be built from generated top-level definition
- no behavioral regressions in current tests

## Validation

Run at minimum:

- `go test ./dsl/...`
- `go test ./...`

Also add focused tests:

- a test that `lex.NewLexer(...)` still tokenizes representative input
- a test that `parse.NewParser(...)` still parses statements into `DefinitionResult`
- a test that building the top-level internal pipeline succeeds

## Risks

### Risk: name/port mismatches hidden by capitalization rules
Example: `Iip` vs `IIP`.

Mitigation:
- normalize port names in `.fbp` files now if needed
- add explicit tests around generated topology names
- prefer exact exported field names where possible

### Risk: package import layering
`dsl` already depends on `lex`, `parse`, and `types`. Keep generated definitions in `dsl/` so they can be consumed from a central place without new circular imports.

---

# Phase 3 — Cache internal factory registration

## Objective

Stop paying the internal factory creation + registration cost on every parse.

## Current issue

In `dsl/api.go`, `createLexerParser()` currently:

- allocates a new `goflow.Factory`
- calls `RegisterComponents(f)` every time

That setup is repeated for every parse request.

## Deliverable

A package-private, once-initialized internal factory for DSL internals.

## Step-by-step

### 3.1 Introduce internal factory accessor
Add to new internal runtime file:

- package-level `sync.Once`
- package-level stored factory
- package-level stored init error

Pattern:

- `internalFactory() (*goflow.Factory, error)`

Responsibilities:
- create factory once
- register DSL components once
- return same instance afterward

### 3.2 Define allowed usage
Document clearly:

- internal factory is initialized once
- after init, it is treated as immutable
- callers must not register/unregister additional components on it

### 3.3 Route all internal graph creation through it
Update all internal constructor paths from phase 2 to use:
- `internalFactory()`

instead of:
- `goflow.NewFactory()`
- `RegisterComponents(...)`

### 3.4 Avoid leaking internal factory publicly
This factory should stay package-private to prevent accidental mutation.

## Files touched

### Add/modify
- `dsl/internal_runtime.go`

### Update callers
- `dsl/api.go`
- possibly `dsl/lex/lexer.go`
- possibly `dsl/parse/parser.go`

## Acceptance criteria

- internal factory registration occurs once per process
- repeated `ParseDefinition(...)` calls no longer rebuild internal registration state
- package behavior remains identical to callers

## Validation

Add tests such as:

- repeated calls to parse definition still succeed
- concurrent calls to `ParseDefinition(...)` do not panic or race
- optional benchmark confirms reduced setup overhead

If you can run with race detector later:
- `go test -race ./dsl/...`

## Risks

### Risk: `Factory` is documented as not concurrency-safe
Mitigation:
- use it in immutable mode after `sync.Once`
- do not mutate after init
- keep usage scoped to `Create(...)`

### Optional future hardening
If desired later in core `goflow`, add a documented “frozen factory” or mutex protection around `Create`. But this blueprint does not require that as a prerequisite.

---

# Phase 4 — Replace manual API glue with one top-level graph

## Objective

Eliminate manual orchestration in `dsl/api.go` and run the whole internal DSL pipeline as a single graph instance.

## Current issue

`dsl/api.go` currently contains a lot of glue logic:

- `createLexerParser`
- `wireLexerInOut`
- `wireParserInOut`
- `runIntermediateStages`
- `runPipeline`

This is both code-heavy and an architectural split: the top-level `.fbp` graph exists, but the runtime path does not actually use it.

## Deliverable

`ParseDefinition(...)` and `LoadDefinitionFile(...)` should use a single internal top-level graph built from generated `dsl/dsl.fbp`.

## Step-by-step

### 4.1 Add top-level graph runner helper
Introduce a helper like:

- `runInternalPipeline(file *types.File) (types.DefinitionResult, error)`

Responsibilities:
- build top-level internal graph from generated definition
- attach `In` and `Out`
- send one `*types.File`
- run graph
- collect one `DefinitionResult`

### 4.2 Simplify `ParseDefinition`
Current:
- special-case empty input
- call `parseFile(...)`

Keep empty-input fast path if desired, but after that:

- create `types.File{Name: "<input>", Data: src}`
- call `runInternalPipeline(...)`
- join `result.Errors` if present
- return `&result.Definition`

### 4.3 Simplify `LoadDefinitionFile`
Keep file reading in place, but after reading:

- create `types.File`
- call `runInternalPipeline(...)`

### 4.4 Remove obsolete glue helpers
Delete or inline obsolete helpers once the top-level graph path is stable:

- `createLexerParser`
- `wireLexerInOut`
- `wireParserInOut`
- `runIntermediateStages`
- `runPipeline`

### 4.5 Preserve empty-input fast path
This one is worth keeping:
- it avoids spinning up the internal graph for empty source
- it is simple and cheap

I’d keep it unless benchmarks show it doesn’t matter.

## Files touched

### Major simplification
- `dsl/api.go`

### Supporting helper
- `dsl/internal_runtime.go` or `dsl/internal_graphs.go`

## Acceptance criteria

- `ParseDefinition(...)` works via one top-level internal graph
- `LoadDefinitionFile(...)` works via one top-level internal graph
- behavior and returned errors remain compatible with current tests
- API-layer code is materially smaller and easier to follow
- top-level `.fbp` file is actually exercised as the runtime topology source

## Validation

Run:
- `go test ./...`

Add focused tests if needed:
- top-level internal pipeline graph emits one `DefinitionResult`
- errors from parser stages still surface correctly through `ParseDefinition`
- empty input still returns an empty `Definition`

## Risks

### Risk: channel closing semantics change
The current implementation manually closes intermediary channels after stage completion.

Mitigation:
- rely on the graph runtime’s existing channel and port semantics
- keep integration tests around EOF, empty input, syntax errors, and one-line files
- add one test specifically for pipeline completion

### Risk: internal graph output waits forever if output not closed
Mitigation:
- use a buffered result channel of size 1 when wiring `Out`
- run the graph exactly once per parse call
- verify the wait channel completes in tests

---

# Cross-phase sequencing

## Recommended order
Do these in strict order:

1. **Phase 1**: generator
2. **Phase 2**: build lexer/parser/top-level from generated definitions
3. **Phase 3**: internal cached factory
4. **Phase 4**: simplify `dsl/api.go` to use one top-level graph

This order minimizes risk because each later phase can reuse earlier scaffolding.

---

# Suggested PR breakdown

## PR 1 — Generator only
Includes:
- `dsl/cmd/dslgen/main.go`
- `dsl/generate.go`
- `dsl/internal_defs_gen.go`

No runtime behavior changes yet.

## PR 2 — Generated internal graph builders
Includes:
- refactor `dsl/lex/lexer.go`
- refactor `dsl/parse/parser.go`
- add internal top-level graph constructor

Behavior should remain unchanged.

## PR 3 — Cached internal factory
Includes:
- `dsl/internal_runtime.go`
- route internal graph construction through cached factory

## PR 4 — API simplification
Includes:
- rewrite `dsl/api.go` around top-level graph runner
- delete manual glue helpers

---

# Test plan by phase

## Existing tests that should remain green
- `dsl/api_test.go`
- `dsl/integration_test.go`
- `dsl/lex/lexer_test.go`
- `dsl/parse/*_test.go`

## New tests to add

### After phase 1
- generated definitions sanity test

### After phase 2
- internal graph constructors build successfully from generated definitions

### After phase 3
- repeated parse calls reuse internal runtime safely

### After phase 4
- top-level pipeline completes and returns exactly one `DefinitionResult`

## Optional benchmarks
Good to add before/after phase 4:

- `BenchmarkParseDefinition_Small`
- `BenchmarkParseDefinition_Medium`
- `BenchmarkParseDefinition_Large`

Even if optimization is not the only goal, these give you proof that setup overhead decreased.

---

# Definition of done

This 4-phase blueprint is complete when:

- `.fbp` assets are the only authored topology source
- generated definitions are checked in
- internal graphs are built from generated definitions, not handwritten wiring
- internal factory registration happens once
- `dsl/api.go` no longer manually orchestrates lexer/parser/intermediate stages
- all tests pass
- package still clearly showcases GoFlow-native implementation

---

# Recommended implementation notes

## Keep public API unchanged
Do not alter:

- `dsl.ParseDefinition`
- `dsl.LoadDefinitionFile`
- `dsl.Parse`
- `dsl.LoadFile`
- `dsl.Build`
- `dsl.UnmarshalDefinition`

## Keep the `.fbp` files visible and documented
They are part of the package’s value proposition. After this refactor, they become even more important.

## Prefer deletion over layering
If a helper is made obsolete by the new architecture, remove it rather than keeping both paths.
