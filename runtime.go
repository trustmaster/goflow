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

func (r *Runtime) runtimeGetRuntime(ws *websocket.Conn, payload interface{}) {
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

func (r *Runtime) graphClear(ws *websocket.Conn, payload interface{}) {
	msg := payload.(clearGraph)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphAddNode(ws *websocket.Conn, payload interface{}) {
	msg := payload.(addNode)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphRemoveNode(ws *websocket.Conn, payload interface{}) {
	msg := payload.(removeNode)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphRenameNode(ws *websocket.Conn, payload interface{}) {
	msg := payload.(renameNode)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphChangeNode(ws *websocket.Conn, payload interface{}) {
	msg := payload.(changeNode)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphAddEdge(ws *websocket.Conn, payload interface{}) {
	msg := payload.(addEdge)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphRemoveEdge(ws *websocket.Conn, payload interface{}) {
	msg := payload.(removeEdge)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphChangeEdge(ws *websocket.Conn, payload interface{}) {
	msg := payload.(changeEdge)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphAddInitial(ws *websocket.Conn, payload interface{}) {
	msg := payload.(addInitial)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphRemoveInitial(ws *websocket.Conn, payload interface{}) {
	msg := payload.(removeInitial)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphAddInPort(ws *websocket.Conn, payload interface{}) {
	msg := payload.(addPort)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphRemoveInPort(ws *websocket.Conn, payload interface{}) {
	msg := payload.(removePort)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphRenameInPort(ws *websocket.Conn, payload interface{}) {
	msg := payload.(renamePort)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphAddOutPort(ws *websocket.Conn, payload interface{}) {
	msg := payload.(addPort)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphRemoveOutPort(ws *websocket.Conn, payload interface{}) {
	msg := payload.(removePort)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) graphRenameOutPort(ws *websocket.Conn, payload interface{}) {
	msg := payload.(renamePort)
	// TODO
	websocket.JSON.Send(ws, msg)
}

func (r *Runtime) componentList(ws *websocket.Conn, payload interface{}) {
	// TODO
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
	r.handlers["runtime.getruntime"] = r.runtimeGetRuntime
	r.handlers["graph.clear"] = r.graphClear
	r.handlers["graph.addnode"] = r.graphAddNode
	r.handlers["graph.removenode"] = r.graphRemoveNode
	r.handlers["graph.renamenode"] = r.graphRenameNode
	r.handlers["graph.changenode"] = r.graphChangeNode
	r.handlers["graph.addedge"] = r.graphAddEdge
	r.handlers["graph.removedge"] = r.graphRemoveEdge
	r.handlers["graph.changeedge"] = r.graphChangeEdge
	r.handlers["graph.addinitial"] = r.graphAddInitial
	r.handlers["graph.removeinitial"] = r.graphRemoveInitial
	r.handlers["graph.addinport"] = r.graphAddInPort
	r.handlers["graph.removeinport"] = r.graphRemoveInPort
	r.handlers["graph.renameinport"] = r.graphRenameInPort
	r.handlers["graph.addoutport"] = r.graphAddOutPort
	r.handlers["graph.removeoutport"] = r.graphRemoveOutPort
	r.handlers["graph.renameoutport"] = r.graphRenameOutPort
	r.handlers["component.list"] = r.componentList
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
