# go-flow - Flow-based programming library for Go (golang) #

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
