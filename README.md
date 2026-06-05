# GoFlow

[![Go Reference](https://pkg.go.dev/badge/github.com/trustmaster/goflow.svg)](https://pkg.go.dev/github.com/trustmaster/goflow)
[![Go Version](https://img.shields.io/github/go-mod/go-version/trustmaster/goflow)](https://golang.org)
[![CI](https://github.com/trustmaster/goflow/actions/workflows/ci.yml/badge.svg)](https://github.com/trustmaster/goflow/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/trustmaster/goflow/branch/master/graph/badge.svg)](https://codecov.io/gh/trustmaster/goflow)

Dataflow and Flow-based programming library for Go.

GoFlow is a lean and opinionated implementation of [Flow-based programming (FBP)](http://en.wikipedia.org/wiki/Flow-based_programming) that lets you design applications as graphs of components reacting to data as it flows through the graph.

> **Flow-based programming** is a programming paradigm that defines applications as networks of black-box processes that exchange data across predefined connections by message passing, where the connections are specified externally to the processes.

## Features

- **Concurrent** — graph nodes run in parallel via goroutines and channels.
- **Structural** — applications are described as components, their ports, and the connections between them.
- **Reactive** — system behavior is defined by how components react to events or manage their lifecycle.
- **Asynchronous by default** — events have no predetermined order unless you enforce one.
- **Isolated** — communication replaces shared state; components don't share memory.

## Installation

GoFlow requires Go 1.23 or later.

```bash
go get github.com/trustmaster/goflow
```

Then import it in your code:

```go
import "github.com/trustmaster/goflow"
```

## Quick Start

Below is a complete program that builds a simple two-component network: one greets names, the other prints the result.

![Greeter example diagram](http://flowbased.wdfiles.com/local--files/goflow/goflow-hello.png)

```go
package main

import (
	"fmt"

	"github.com/trustmaster/goflow"
)

// Greeter sends greetings.
type Greeter struct {
	Name <-chan string // input port
	Res  chan<- string // output port
}

// Process reads incoming names and sends back greetings.
func (c *Greeter) Process() {
	for name := range c.Name {
		c.Res <- fmt.Sprintf("Hello, %s!", name)
	}
}

// Printer prints its input on screen.
type Printer struct {
	Line <-chan string // input port
}

// Process reads lines and prints them.
func (c *Printer) Process() {
	for line := range c.Line {
		fmt.Println(line)
	}
}

// NewGreetingApp defines the application graph.
func NewGreetingApp() *goflow.Graph {
	n := goflow.NewGraph()
	n.Add("greeter", new(Greeter))
	n.Add("printer", new(Printer))
	n.Connect("greeter", "Res", "printer", "Line")
	n.MapInPort("In", "greeter", "Name")
	return n
}

func main() {
	net := NewGreetingApp()
	in := make(chan string)
	net.SetInPort("In", in)

	wait := goflow.Run(net)

	in <- "John"
	in <- "Boris"
	in <- "Hanna"
	close(in)

	<-wait
}
```

> **Why the `wait` channel?** The flow-based world is asynchronous — events don't necessarily happen in the order they were sent. The `wait` channel signals when the network has fully completed, preventing premature program termination.

## Terminology

| Term      | Description |
|-----------|-------------|
| **Component** | The basic processing element. Its structure consists of input/output ports and state fields; its behavior is defined by event handlers. Analogous to a Class. |
| **Connection** | A link between two ports in the graph. In GoFlow this is a typed channel. |
| **Graph** | A higher-level entity composed of components and connections. Can represent composite components or entire applications. Analogous to a Class. |
| **Network** | A running instance of a Graph. Analogous to an Object. |
| **Port** | A property through which a Component or Graph communicates with the outside world (input/output). In GoFlow this is a channel field. |
| **Process** | A running instance of a Component. Analogous to an Object. |

More terms can be found in the [Flowbased.org Terminology](https://github.com/flowbased/flowbased.org/wiki/Terminology) and the [FBP wiki](http://www.jpaulmorrison.com/cgi-bin/wiki.pl?action=index).

## DSL / FBP Syntax

GoFlow includes a built-in **`dsl`** package that lets you define application graphs using standard [Flow-Based Programming](http://en.wikipedia.org/wiki/Flow-based_programming) (FBP) syntax in `.fbp` files instead of writing Go code.

```go
import "github.com/trustmaster/goflow/dsl"
```

A minimal `.fbp` file looks like this:

```fbp
# Declare the sender's input as a graph port
INPORT=Sender.IN:IN
# Declare the receiver's output as a graph port
OUTPORT=Receiver.OUT:OUT

# Declare two processes and connect them
Sender(test/sender) OUT -> IN Receiver(test/receiver)
```

Load and run it with:

```go
g, err := dsl.LoadFile("graph.fbp", factory)
```

The `dsl` package supports process declarations, connections, array ports, IIPs (initial information packets), and port exports. See the [`dsl` package README](dsl/README.md) for the full syntax guide and API reference.

## Documentation

### Wiki

- [Components](https://github.com/trustmaster/goflow/wiki/Components) — ports, events, process, and state.
- [Graphs](https://github.com/trustmaster/goflow/wiki/Graphs) — structure definition and behavior.

### GoDoc

```bash
go doc github.com/trustmaster/goflow
```

Or view the [online reference](https://pkg.go.dev/github.com/trustmaster/goflow).

## Related Projects

- [Flow-based.org](https://github.com/flowbased/flowbased.org/wiki) — specifications and recommendations for FBP systems.
- [J. Paul Morrison's Flow-Based Programming](https://jpaulm.github.io/fbp/index.html) — the origin of FBP, including [JavaFBP](https://github.com/jpaulm/javafbp), [C#FBP](https://github.com/jpaulm/csharpfbp), and the [DrawFBP](https://github.com/jpaulm/drawfbp) diagramming tool.
- [NoFlo](http://noflojs.org/) — FBP for JavaScript and Node.js.

## Roadmap

- Integration with NoFlo-UI / Flowhub (in progress)
- Distributed networks via TCP/IP and UDP
- Reflection and monitoring of networks

## Contributing

Contributions are welcome! Please open an issue or pull request on [GitHub](https://github.com/trustmaster/goflow).

Before submitting changes, make sure your code passes the linter and tests:

```bash
make test
```

## License

[MIT](LICENSE)
