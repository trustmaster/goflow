# GoFlow - Flow-based programming library for Go (golang) #

This is quite a minimalistic implementation of [Flow-based programming](http://en.wikipedia.org/wiki/Flow-based_programming) in Go programming language that aims at designing applications as graphs of components which react to data that flows through the graph.

The main properties of the proposed model are:

* Concurrent - graph nodes run in parallel.
* Structural - applications are described as components, their ports and connections between them.
* Event-driven - system's behavior is how components react to events.
* Asynchronous - there is no determined order in which events happen.
* Isolated - sharing is done by communication, state is not shared.

## Getting started ##

Current version of the library requires at least Go weekly.2012-01-27 and aims at compatibility with upcoming Go 1 release. If you don't have the Go compiler installed, read the official [Go install guide](http://golang.org/doc/install.html) or if you use Ubuntu read [Ubuntu Go Wiki](https://wiki.ubuntu.com/Go).

Use go tool to install the package in your packages tree:

```
go get github.com/trustmaster/goflow
```

Then you can use it in import section of your Go programs:

```go
import "github.com/trustmaster/goflow"
```

## Basic Example ##

Below there is a listing of a simple program running a network of two processes.

![Greeter example diagram](http://flowbased.wdfiles.com/local--files/goflow/goflow-hello.png)

This first one generates greetings for given names, the second one prints them on screen. It demonstrates how components and graphs are defined and how they are embedded into the main program.

```go
package main

import (
	"fmt"
	"github.com/trustmaster/goflow"
)

// A component that generates greetings
type Greeter struct {
	flow.Component               // component "superclass" embedded
	Name           <-chan string // input port
	Res            chan<- string // output port
}

// Reaction to a new name input
func (g *Greeter) OnName(name string) {
	greeting := fmt.Sprintf("Hello, %s!", name)
	// send the greeting to the output port
	g.Res <- greeting
}

// A component that prints its input on screen
type Printer struct {
	flow.Component
	Line <-chan string // inport
}

// Prints a line when it gets it
func (p *Printer) OnLine(line string) {
	fmt.Println(line)
}

// Our greeting network
type GreetingApp struct {
	flow.Graph // graph "superclass" embedded
}

// Graph constructor and structure definition
func NewGreetingApp() *GreetingApp {
	n := new(GreetingApp) // creates the object in heap
	n.InitGraphState()    // allocates memory for the graph
	// Add processes to the network
	n.Add(new(Greeter), "greeter")
	n.Add(new(Printer), "printer")
	// Connect them with a channel
	n.Connect("greeter", "Res", "printer", "Line", make(chan string))
	// Our net has 1 inport mapped to greeter.Name
	n.MapInPort("In", "greeter", "Name")
	return n
}

// We need this flag to terminate when processing is finished
var finish chan bool

// Use this handler to let the main() know when the network terminates
func (a *GreetingApp) Finish() {
	finish <- true
}

func main() {
	// Termination signal channel
	finish = make(chan bool)
	// Create the network
	net := NewGreetingApp()
	// We need a channel to talk to it
	in := make(chan string)
	net.SetInPort("In", in)
	// Run the net
	flow.RunNet(net)
	// Now we can send some names and see what happens
	in <- "John"
	in <- "Boris"
	in <- "Hanna"
	// Close the input to shut the network down
	close(in)
	// Wait until it's done
	<-finish
}
```

Looks a bit heavy for such a simple task but FBP is aimed at a bit more complex things than just printing on screen. So in more complex an realistic examples the infractructure pays the price.

You probably have one question left even after reading the comments in code: why do we need to wait for the finish signal? This is because flow-based world is asynchronous and while you expect things to happen in the same sequence as they are in main(), during runtime they don't necessarily follow the same order and the application might terminate before the network has done its job. To avoid this confusion we added the finish signal for synchronization.

## Terminology ##

Here are some Flow-based programming terms used in GoFlow:

* Component - the basic element that processes data. Its structure consists of input and output ports and state fields. Its behavior is the set of event handlers. In OOP terms Component is a Class.
* Connection - a link between 2 ports in the graph. In Go it is a channel of specific type.
* Graph - components and connections between them, forming a higher level entity. Graphs may represent composite components or entire applications. In OOP terms Graph is a Class.
* Network - is a Graph instance running in memory. In OOP terms a Network is an object of Graph class.
* Port - is a property of a Component or Graph through which it communicates with the outer world. There are input ports (Inports) and output ports (Outports). For GoFlow components it is a channel field.
* Process - is a Component instance running in memory. In OOP terms a Process is an object of Component class.

More terms can be found in [flowbased terms](http://flowbased.org/terms) and [FBP wiki](http://www.jpaulmorrison.com/cgi-bin/wiki.pl?action=index).

## Further reading ##

For more information please visit [project wiki](https://github.com/trustmaster/goflow/wiki).

Documentation for the flow package can be accessed using standard godoc tool, e.g.

```
godoc github.com/trustmaster/goflow
```

## More examples ##

* [GoChat](https://github.com/trustmaster/gochat), a simple chat in Go using this library

## Links ##

Here are related projects and resources:

* [J. Paul Morrison's Flow-Based Programming](http://www.jpaulmorrison.com/fbp/), the origin of FBP, [JavaFBP, C#FBP](http://sourceforge.net/projects/flow-based-pgmg/) and [DrawFBP](http://www.jpaulmorrison.com/fbp/#DrawFBP) diagramming tool.
* [Knol about FBP](http://knol.google.com/k/flow-based-programming)
* [NoFlo](https://github.com/bergie/noflo), FBP for Node.js
* [Pypes](http://www.pypes.org/), flow-based Python ETL
* [Go](http://golang.org/), the Go programming language

## TODO ##

* A tool to convert visual diagrams into Go code
* Distributed networks via TCP/IP and UDP
* Better run-time restructuring and evolution
* Reflection and monitoring of networks
