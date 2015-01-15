# GoFlow - Dataflow and Flow-based programming library for Go (golang)

This is quite a minimalistic implementation of [Flow-based programming](http://en.wikipedia.org/wiki/Flow-based_programming) and several other concurrent models in Go programming language that aims at designing applications as graphs of components which react to data that flows through the graph.

The main properties of the proposed model are:

* Concurrent - graph nodes run in parallel.
* Structural - applications are described as components, their ports and connections between them.
* Event-driven - system's behavior is how components react to events.
* Asynchronous - there is no determined order in which events happen.
* Isolated - sharing is done by communication, state is not shared.

## Getting started

Current version of the library requires a latest stable Go release. If you don't have the Go compiler installed, read the official [Go install guide](http://golang.org/doc/install).

Use go tool to install the package in your packages tree:

```
go get github.com/Synthace/goflow
```

Then you can use it in import section of your Go programs:

```go
import "github.com/Synthace/goflow"
```

## Basic Example

Below there is a listing of a simple program running a network of two processes.

![Greeter example diagram](http://flowbased.wdfiles.com/local--files/goflow/goflow-hello.png)

This first one generates greetings for given names, the second one prints them on screen. It demonstrates how components and graphs are defined and how they are embedded into the main program.

```go
package main

import (
	"fmt"
	"github.com/Synthace/goflow"
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
	flow.Graph               // graph "superclass" embedded
}

// Graph constructor and structure definition
func NewGreetingApp() *GreetingApp {
	n := new(GreetingApp) // creates the object in heap
	n.InitGraphState()    // allocates memory for the graph
	// Add processes to the network
	n.Add(new(Greeter), "greeter")
	n.Add(new(Printer), "printer")
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
	flow.RunNet(net)
	// Now we can send some names and see what happens
	in <- "John"
	in <- "Boris"
	in <- "Hanna"
	// Close the input to shut the network down
	close(in)
	// Wait until the app has done its job
	<-net.Wait()
}
```

Looks a bit heavy for such a simple task but FBP is aimed at a bit more complex things than just printing on screen. So in more complex an realistic examples the infractructure pays the price.

You probably have one question left even after reading the comments in code: why do we need to wait for the finish signal? This is because flow-based world is asynchronous and while you expect things to happen in the same sequence as they are in main(), during runtime they don't necessarily follow the same order and the application might terminate before the network has done its job. To avoid this confusion we listen for a signal on network's `Wait()` channel which is closed when the network finishes its job.

## Basic Server for Flowhub
```go
package main

import (
    "github.com/Synthace/goflow"
    "fmt"
)

//Call everything from main function
func main() {
    port := ":3569"
    fmt.Println(port)
    runtime := new(flow.Runtime)
    runtime.Init()
    fmt.Println(runtime)
    runtime.Listen(port)
}
```

## Terminology

Here are some Flow-based programming terms used in GoFlow:

* Component - the basic element that processes data. Its structure consists of input and output ports and state fields. Its behavior is the set of event handlers. In OOP terms Component is a Class.
* Connection - a link between 2 ports in the graph. In Go it is a channel of specific type.
* Graph - components and connections between them, forming a higher level entity. Graphs may represent composite components or entire applications. In OOP terms Graph is a Class.
* Network - is a Graph instance running in memory. In OOP terms a Network is an object of Graph class.
* Port - is a property of a Component or Graph through which it communicates with the outer world. There are input ports (Inports) and output ports (Outports). For GoFlow components it is a channel field.
* Process - is a Component instance running in memory. In OOP terms a Process is an object of Component class.

More terms can be found in [flowbased terms](http://flowbased.org/terms) and [FBP wiki](http://www.jpaulmorrison.com/cgi-bin/wiki.pl?action=index).

## Documentation

### Contents

1. [Components](https://github.com/trustmaster/goflow/wiki/Components)
    1. [Ports, Events and Handlers](https://github.com/trustmaster/goflow/wiki/Components#ports-events-and-handlers)
    2. [Processes and their lifetime](https://github.com/trustmaster/goflow/wiki/Components#processes-and-their-lifetime)
    3. [State](https://github.com/trustmaster/goflow/wiki/Components#state)
    4. [Concurrency](https://github.com/trustmaster/goflow/wiki/Components#concurrency)
    5. [Internal state and Thread-safety](https://github.com/trustmaster/goflow/wiki/Components#internal-state-and-thread-safety)
2. [Graphs](https://github.com/trustmaster/goflow/wiki/Graphs)
    1. [Structure definition](https://github.com/trustmaster/goflow/wiki/Graphs#structure-definition)
    2. [Behavior](https://github.com/trustmaster/goflow/wiki/Graphs#behavior)

### Package docs

Documentation for the flow package can be accessed using standard godoc tool, e.g.

```
godoc github.com/trustmaster/goflow
```

## Runtime registration example
```go
package main

import (
    "fmt"
	"log"
	"github.com/nu7hatch/gouuid"
    "encoding/json"
    "strings"
    "github.com/parnurzeal/gorequest"
)

// Runtime_details stores data needed to register a runtime.
// Variable names must be capitalized for two reasons: 1. Visibility outside of package.
// 2. Flowhub requires "type" as a passed parameter, but "type" is a reserved word in Go.
// This struct is json-ified using json.Marshal and made lowercase using strings.ToLower
type Runtime_details struct {
	Type string
    Protocol string
    Address string
	Id string
	Label string
    Port string
	User string
}

// Initialize Runtime_details with user-provided parameters.  Runtime ID is automatically
// generated with uuid.
func (r *Runtime_details) runtimeInitializeRuntime(runtime_type string , protocol string , user_id string , label string , ip string , port string){
	uv4, err := uuid.NewV4()
	if err != nil {
		log.Println(err.Error())
	}
	s := []string{ip, port}
    
	r.Type = runtime_type
    r.Protocol = protocol
    r.Address = strings.Join(s,"")
	r.Id = uv4.String()
	r.Label = label
    r.Port = port
    r.User = user_id
}

func main() {
    
    newRuntime := new(Runtime_details)
    newRuntimeP := &newRuntime
    newRuntimeP.runtimeInitializeRuntime("fbp-go-example", "websocket", "<user id here>", "go runtime test", "ws://localhost:", "3569")
    
    newRuntimeJSON, err := json.Marshal(newRuntime)
	if err != nil {
		log.Println(err.Error())
	}
    
    stringifiedNewRuntimeJSON := string(newRuntimeJSON[:])
    stringifiedNewRuntimeJSON = strings.ToLower(stringifiedNewRuntimeJSON)
    fmt.Println("stringifiedNewRuntimeJSON:", stringifiedNewRuntimeJSON)
    
	s := []string{"http://api.flowhub.io/runtimes/", newRuntime.Id}
    url := strings.Join(s,"")
    fmt.Println("url: ",url)
    
    request := gorequest.New()
    resp, body, errs := request.Put(url).
    Set("Content-Type", "application/json").
    Send(stringifiedNewRuntimeJSON).
    End()
    
    fmt.Println("resp", resp)
    fmt.Println("body", body)
    fmt.Println("errs", errs)
    
}
```
## More examples

* [GoChat](https://github.com/trustmaster/gochat), a simple chat in Go using this library

## Links

Here are related projects and resources:

* [J. Paul Morrison's Flow-Based Programming](http://www.jpaulmorrison.com/fbp/), the origin of FBP, [JavaFBP, C#FBP](http://sourceforge.net/projects/flow-based-pgmg/) and [DrawFBP](http://www.jpaulmorrison.com/fbp/#DrawFBP) diagramming tool.
* [Knol about FBP](http://knol.google.com/k/flow-based-programming)
* [NoFlo](http://noflojs.org/), FBP for JavaScript and Node.js
* [Pypes](http://www.pypes.org/), flow-based Python ETL
* [Go](http://golang.org/), the Go programming language

## TODO

* Integration with NoFlo-UI
* Distributed networks via TCP/IP and UDP
* Better run-time restructuring and evolution
* Reflection and monitoring of networks
