package flow

import (
	"code.google.com/p/go.net/websocket"
	"github.com/nu7hatch/gouuid"
	"log"
	"net"
	"net/http"
)

type protocolHandler func(*websocket.Conn, interface{})

// Runtime is a NoFlo-compatible runtime implementing the FBP protocol
type Runtime struct {
	// Unique runtime ID for use with Flowhub
	id string
	// Protocol command handlers
	handlers map[string]protocolHandler
	// Graphs created at runtime and exposed as components
	graphs map[string]*Graph
	// Main graph ID
	mainId string
	// Main graph
	main *Graph
	// Websocket server onReady signal
	ready chan struct{}
	// Websocket server onShutdown signal
	done chan struct{}
}

// Register command handlers
func (r *Runtime) Init() {
	uv4, err := uuid.NewV4()
	if err != nil {
		log.Println(err.Error())
	}
	r.id = uv4.String()
	r.done = make(chan struct{})
	r.ready = make(chan struct{})
	r.handlers = make(map[string]protocolHandler)
	r.handlers["runtime.getruntime"] = func(ws *websocket.Conn, payload interface{}) {
		websocket.JSON.Send(ws, runtimeInfo{"goflow",
			"0.4",
			[]string{"protocol:runtime",
				"protocol:graph",
				"protocol:component",
				"protocol:network",
				"component:getsource"},
			r.id,
		})
	}
	r.handlers["graph.clear"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(clearGraph)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.addnode"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(addNode)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.removenode"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(removeNode)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.renamenode"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(renameNode)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.changenode"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(changeNode)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.addedge"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(addEdge)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.removedge"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(removeEdge)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.changeedge"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(changeEdge)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.addinitial"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(addInitial)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.removeinitial"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(removeInitial)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.addinport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(addPort)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.removeinport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(removePort)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.renameinport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(renamePort)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.addoutport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(addPort)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.removeoutport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(removePort)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["graph.renameoutport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(renamePort)
		// TODO
		websocket.JSON.Send(ws, msg)
	}
	r.handlers["component.list"] = func(ws *websocket.Conn, payload interface{}) {
		// TODO
	}
}

// Id returns runtime's UUID v4
func (r *Runtime) Id() string {
	return r.id
}

// Ready returns a channel which is closed when the runtime is ready to work
func (r *Runtime) Ready() chan struct{} {
	return r.ready
}

// Stop tells the runtime to shut down
func (r *Runtime) Stop() {
	close(r.done)
}

func (r *Runtime) Handle(ws *websocket.Conn) {
	defer func() {
		err := ws.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}()
	var msg Message
	if err := websocket.JSON.Receive(ws, &msg); err != nil {
		log.Println(err.Error())
		return
	}
	handler, exists := r.handlers[msg.Protocol+"."+msg.Command]
	if !exists {
		log.Printf("Unknown command: %s.%s\n", msg.Protocol, msg.Command)
		return
	}
	handler(ws, msg.Payload)
}

func (r *Runtime) Listen(address string) {
	http.Handle("/", websocket.Handler(r.Handle))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err.Error())
	}

	go func() {
		err = http.Serve(listener, nil)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}()
	close(r.ready)

	// Wait for termination signal
	<-r.done
	listener.Close()
}
