package flow

import (
	"code.google.com/p/go.net/websocket"
	"github.com/nu7hatch/gouuid"
	"log"
	//"net"
	//"net/http"
    "fmt"
)

type nodeHandler func(*websocket.Conn, interface{})

// Runtime is a NoFlo-compatible runtime implementing the FBP protocol
type Node struct {
	id       string
	handlers map[string]protocolHandler
	ready    chan struct{}
	done     chan struct{}
}

// Register command handlers
func (r *Runtime) anotherInit() {
	uv4, err := uuid.NewV4()
	if err != nil {
		log.Println(err.Error())
	}
	r.id = uv4.String()
	r.done = make(chan struct{})
	r.ready = make(chan struct{})
	r.handlers = make(map[string]protocolHandler)
	r.handlers["runtime.getruntime"] = r.runtimeGetRuntime
	r.handlers["network.getstatus"] = r.networkGetStatus
	r.handlers["component.list"] = r.componentList
	r.handlers["graph.addnode"] = r.graphAddnode
}

func (r *Runtime) graphAddnode(ws *websocket.Conn, payload interface{}) {
    fmt.Println("handle graph.addnode")
    //placeholder
}

func (r *Runtime) graphAddinitial(ws *websocket.Conn, payload interface{}) {
    fmt.Println("handle graph.addinitial")
    //placeholder
}

func (r *Runtime) graphChangenode(ws *websocket.Conn, payload interface{}) {
    fmt.Println("handle graph.changenode")
    //placeholder
}

func (r *Runtime) graphAddedge(ws *websocket.Conn, payload interface{}) {
    fmt.Println("handle graph.addedge")
    //placeholder
}

func (r *Runtime) graphChangeedge(ws *websocket.Conn, payload interface{}) {
    fmt.Println("handle graph.changeedge")
    //placeholder
}

func (r *Runtime) graphRemoveedge(ws *websocket.Conn, payload interface{}) {
    fmt.Println("handle graph.removeedge")
    //placeholder
}

func (r *Runtime) graphRemoveinitial(ws *websocket.Conn, payload interface{}) {
    fmt.Println("handle graph.removeinitial")
    //placeholder
}

func (r *Runtime) graphRemovenode(ws *websocket.Conn, payload interface{}) {
    fmt.Println("handle graph.removenode")
    //placeholder
}