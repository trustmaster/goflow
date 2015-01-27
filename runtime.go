package flow

import (
	"code.google.com/p/go.net/websocket"
	"github.com/nu7hatch/gouuid"
	"log"
	//"net"
	"net/http"
    "fmt"
)

// Message represents a single FBP protocol message
type Message struct {
	// Protocol is NoFlo protocol identifier:
	// "runtime", "component", "graph" or "network"
	Protocol string
	// Command is a command to be executed within the protocol
	Command string
	// Payload is JSON-encoded body of the message
	Payload interface{}
}

type protocolHandler func(*websocket.Conn, interface{})

// Runtime is a NoFlo-compatible runtime implementing the FBP protocol
type Runtime struct {
	id       string
	handlers map[string]protocolHandler
	ready    chan struct{}
	done     chan struct{}
}

//
type wsSend struct {
    Command string `json:"command"`
    Protocol string `json:"protocol"`
    Payload interface{} `json:"payload"`
}

// runtimeInfo message contains response to runtime.getruntime request
type runtimeInfo struct {
	Version      string `json:"version"`
	Type         string `json:"type"`
	Capabilities []string `json:"capabilities"`
	//Id           string `json:"id"`
}

// runtimeInfo message contains response to runtime.getruntime request
type networkInfo struct {
	Graph       string `json:"graph"`
	Running     bool `json:"running"`
	Info        bool `json:"info"`
}

func (r *Runtime) runtimeGetRuntime(ws *websocket.Conn, payload interface{}) {
    fmt.Println("handle runtime.getruntime")
    websocket.JSON.Send(ws, wsSend{"runtime", "runtime", runtimeInfo{"0.4",
        "fbp-go-example",
		[]string{"protocol:runtime",
			//"protocol:graph",
			"protocol:component",
			//"protocol:network",
			//"component:getsource",
            },
		//r.id,
	}})
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
	r.handlers["network.getstatus"] = r.networkGetStatus
	r.handlers["component.list"] = r.componentList
	r.handlers["graph.addnode"] = r.graphAddnode //start here
	r.handlers["graph.addinitial"] = r.graphAddinitial
	r.handlers["graph.changenode"] = r.graphChangenode
	r.handlers["graph.addedge"] = r.graphAddedge
	r.handlers["graph.changeedge"] = r.graphChangeedge
	r.handlers["graph.removeedge"] = r.graphRemoveedge
	r.handlers["graph.removeinitial"] = r.graphRemoveinitial
	r.handlers["graph.removenode"] = r.graphRemovenode
	r.handlers["network.start"] = r.networkStart
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
    fmt.Println("Handling")
	defer func() {
		err := ws.Close()
        fmt.Println("Closed Connection")
		if err != nil {
			log.Println(err.Error())
		}
	}()
    //defer ws.Close();
    for {
        var msg Message
        if err := websocket.JSON.Receive(ws, &msg); err != nil {
            log.Println(err.Error())
            return
        }
        fmt.Println(msg)
        handler, exists := r.handlers[msg.Protocol+"."+msg.Command]
        fmt.Println(msg.Protocol+"."+msg.Command)
        if !exists {
            log.Printf("Unknown command: %s.%s\n", msg.Protocol, msg.Command)
            return
        }
        handler(ws, msg.Payload)
        fmt.Println("Handle Done")
    }
}

func (r *Runtime) Listen(address string) {
	http.Handle("/", websocket.Handler(r.Handle))
    if err := http.ListenAndServe(address, nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
    fmt.Println("End")
    /*fmt.Println("listening thing")
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err.Error())
	}
    fmt.Println("done listening")

	go func() {
		err = http.Serve(listener, nil)
		if err != nil {
			log.Fatalln(err.Error())
		}
	}()
	close(r.ready)
    fmt.Println("ready is closed")

	// Wait for termination signal
	<-r.done
    fmt.Println("runtime is done")
	listener.Close()
    fmt.Println("listener closed")*/
}
