# GoFlow - Dataflow and Flow-based programming library for Go (golang)

[![Build Status](https://travis-ci.com/trustmaster/goflow.svg?branch=master)](https://travis-ci.com/trustmaster/goflow) [![codecov](https://codecov.io/gh/trustmaster/goflow/branch/master/graph/badge.svg)](https://codecov.io/gh/trustmaster/goflow)


### _Status of this branch (WIP)_

_Warning: you are currently on v1 branch of GoFlow. v1 is a revisit and refactoring of the original GoFlow code which remained almost unchanged for 7 years. This branch is deep **in progress**, no stability guaranteed. API also may change._

- _[More information on v1](https://github.com/trustmaster/goflow/issues/49)_
- _[Take me back to v0](https://github.com/trustmaster/goflow/tree/v0)_

_If your code depends on the old implementation, you can build it using [release 0.1](https://github.com/trustmaster/goflow/releases/tag/0.1)._

--

GoFlow is a lean and opinionated implementation of [Flow-based programming](http://en.wikipedia.org/wiki/Flow-based_programming) in Go that aims at designing applications as graphs of components which react to data that flows through the graph.

The main properties of the proposed model are:

* Concurrent - graph nodes run in parallel.
* Structural - applications are described as components, their ports and connections between them.
* Reactive/active - system's behavior is how components react to events or how they handle their lifecycle.
* Asynchronous/synchronous - there is no determined order in which events happen, unless you demand for such order.
* Isolated - sharing is done by communication, state is not shared.

## Getting started

If you don't have the Go compiler installed, read the official [Go install guide](http://golang.org/doc/install).

Use go tool to install the package in your packages tree:

```
go get github.com/trustmaster/goflow
```

Then you can use it in import section of your Go programs:

```go
import "github.com/trustmaster/goflow"
```

## Basic Example

Below there is a listing of a simple program running a network of two processes.

![Greeter example diagram](http://flowbased.wdfiles.com/local--files/goflow/goflow-hello.png)

This first one generates greetings for given names, the second one prints them on screen. It demonstrates how components and graphs are defined and how they are embedded into the main program.

```go
package main

import (
	"fmt"
	"github.com/trustmaster/goflow"
)

// Greeter sends greetings
type Greeter struct {
	Name           <-chan string // input port
	Res            chan<- string // output port
}

// Process incoming data
func (c *Greeter) Process() {
	// Keep reading incoming packets
	for name := range c.Name {
		greeting := fmt.Sprintf("Hello, %s!", name)
		// Send the greeting to the output port
		c.Res <- greeting
	}
}

// Printer prints its input on screen
type Printer struct {
	Line <-chan string // inport
}

// Process prints a line when it gets it
func (c *Printer) Process() {
	for line := range c.Line {
		fmt.Println(line)
	}
}

// NewGreetingApp defines the app graph
func NewGreetingApp() *goflow.Graph {
	n := goflow.NewGraph()
	// Add processes to the network
	n.Add("greeter", new(Greeter))
	n.Add("printer", new(Printer))
	// Connect them with a channel
	n.Connect("greeter", "Res", "printer", "Line")
	// Our net has 1 inport mapped to greeter.Name
	n.MapInPort("In", "greeter", "Name")
	return n
}

func main() {
	// Create the network
	net := NewGreetingApp()
	// We need a channel to talk to it
	in := make(chan string)
	net.SetInPort("In", in)
	// Run the net
	wait := goflow.Run(net)
	// Now we can send some names and see what happens
	in <- "John"
	in <- "Boris"
	in <- "Hanna"
	// Send end of input
	close(in)
	// Wait until the net has completed its job
	<-wait
}
```

Looks a bit heavy for such a simple task but FBP is aimed at a bit more complex things than just printing on screen. So in more complex an realistic examples the infractructure pays the price.

You probably have one question left even after reading the comments in code: why do we need to wait for the finish signal? This is because flow-based world is asynchronous and while you expect things to happen in the same sequence as they are in main(), during runtime they don't necessarily follow the same order and the application might terminate before the network has done its job. To avoid this confusion we listen for a signal on network's `wait` channel which is sent when the network finishes its job.

## Terminology

Here are some Flow-based programming terms used in GoFlow:

* Component - the basic element that processes data. Its structure consists of input and output ports and state fields. Its behavior is the set of event handlers. In OOP terms Component is a Class.
* Connection - a link between 2 ports in the graph. In Go it is a channel of specific type.
* Graph - components and connections between them, forming a higher level entity. Graphs may represent composite components or entire applications. In OOP terms Graph is a Class.
* Network - is a Graph instance running in memory. In OOP terms a Network is an object of Graph class.
* Port - is a property of a Component or Graph through which it communicates with the outer world. There are input ports (Inports) and output ports (Outports). For GoFlow components it is a channel field.
* Process - is a Component instance running in memory. In OOP terms a Process is an object of Component class.

More terms can be found in [Flow-based Wiki Terms](https://github.com/flowbased/flowbased.org/wiki/Terminology) and [FBP wiki](http://www.jpaulmorrison.com/cgi-bin/wiki.pl?action=index).

## Documentation

### Contents

1. [Components](https://github.com/trustmaster/goflow/wiki/Components)
    1. [Ports and Events](https://github.com/trustmaster/goflow/wiki/Components#ports-and-events)
    2. [Process](https://github.com/trustmaster/goflow/wiki/Components#process)
    3. [State](https://github.com/trustmaster/goflow/wiki/Components#state)
2. [Graphs](https://github.com/trustmaster/goflow/wiki/Graphs)
    1. [Structure definition](https://github.com/trustmaster/goflow/wiki/Graphs#structure-definition)
    2. [Behavior](https://github.com/trustmaster/goflow/wiki/Graphs#behavior)

### Package docs

Documentation for the flow package can be accessed using standard godoc tool, e.g.

```
godoc github.com/trustmaster/goflow
```

## Links

Here are related projects and resources:

* [Flowbased.org](https://github.com/flowbased/flowbased.org/wiki), specifications and recommendations for FBP systems.
* [J. Paul Morrison's Flow-Based Programming](https://jpaulm.github.io/fbp/index.html), the origin of FBP, [JavaFBP](https://github.com/jpaulm/javafbp), [C#FBP](https://github.com/jpaulm/csharpfbp) and [DrawFBP](https://github.com/jpaulm/drawfbp) diagramming tool.
* [NoFlo](http://noflojs.org/), FBP for JavaScript and Node.js
* [Go](http://golang.org/), the Go programming language

## TODO

* Integration with NoFlo-UI/Flowhub (in progress)
* Distributed networks via TCP/IP and UDP
* Reflection and monitoring of networks
