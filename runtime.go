package flow

import (
	"github.com/gorilla/websocket"
	"github.com/nu7hatch/gouuid"
	"log"
	"net/http"
	"encoding/json"
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
	// Gorilla Webscocket upgrader
	upgrader websocket.Upgrader
}

func sendJSON(ws *websocket.Conn, msg interface{}) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Println("JSON encoding error", err)
		return
	}
	err = ws.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		log.Println("Websocket write error", err)
	}
}

// Register command handlers
func (r *Runtime) Init(name string) {
	uv4, err := uuid.NewV4()
	if err != nil {
		log.Println(err.Error())
	}
	r.id = uv4.String()
	r.done = make(chan struct{})
	r.ready = make(chan struct{})
	r.handlers = make(map[string]protocolHandler)
	r.handlers["runtime.getruntime"] = func(ws *websocket.Conn, payload interface{}) {
		sendJSON(ws, runtimeMessage{
			Protocol: "runtime",
			Command:  "runtime",
			Payload: runtimeInfo{Type: name,
				Version: "0.4",
				Capabilities: []string{"protocol:runtime",
					"protocol:graph",
					"protocol:component",
					"protocol:network",
					"component:getsource"},
				Id: r.id,
			},
		})
	}
	r.handlers["graph.clear"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(clearGraph)
		r.graphs[msg.Id] = new(Graph)
		r.graphs[msg.Id].InitGraphState()
		if msg.Main {
			r.mainId = msg.Id
			r.main = r.graphs[msg.Id]
		}
		if _, exists := ComponentRegistry[msg.Id]; !exists {
			Register(msg.Id, func() interface{} {
				net := new(Graph)
				net.InitGraphState()
				return net
			})
		}
		Annotate(msg.Id, ComponentInfo{
			Description: msg.Description,
			Icon:        msg.Icon,
		})
		UpdateComponentInfo(msg.Id)
		entry, _ := ComponentRegistry[msg.Id]
		sendJSON(ws, componentMessage{
			Protocol: "component",
			Command:  "component",
			Payload:  entry.Info,
		})
	}
	r.handlers["graph.addnode"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(addNode)
		r.graphs[msg.Graph].AddNew(msg.Component, msg.Id)
	}
	r.handlers["graph.removenode"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(removeNode)
		r.graphs[msg.Graph].Remove(msg.Id)
	}
	r.handlers["graph.renamenode"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(renameNode)
		r.graphs[msg.Graph].Rename(msg.From, msg.To)
	}
	r.handlers["graph.changenode"] = func(ws *websocket.Conn, payload interface{}) {
		// Currently unsupported
	}
	r.handlers["graph.addedge"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(addEdge)
		r.graphs[msg.Graph].Connect(msg.Src.Node, msg.Src.Port, msg.Tgt.Node, msg.Tgt.Port)
	}
	r.handlers["graph.removedge"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(removeEdge)
		r.graphs[msg.Graph].Disconnect(msg.Src.Node, msg.Src.Port, msg.Tgt.Node, msg.Tgt.Port)
	}
	r.handlers["graph.changeedge"] = func(ws *websocket.Conn, payload interface{}) {
		// Currently unsupported
	}
	r.handlers["graph.addinitial"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(addInitial)
		r.graphs[msg.Graph].AddIIP(msg.Src.Data, msg.Tgt.Node, msg.Tgt.Port)
	}
	r.handlers["graph.removeinitial"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(removeInitial)
		r.graphs[msg.Graph].RemoveIIP(msg.Tgt.Node, msg.Tgt.Port)
	}
	r.handlers["graph.addinport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(addPort)
		r.graphs[msg.Graph].MapInPort(msg.Public, msg.Node, msg.Port)
		UpdateComponentInfo(msg.Graph)
		entry, _ := ComponentRegistry[msg.Graph]
		sendJSON(ws, componentMessage{
			Protocol: "component",
			Command:  "component",
			Payload:  entry.Info,
		})
	}
	r.handlers["graph.removeinport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(removePort)
		r.graphs[msg.Graph].UnsetInPort(msg.Public)
		r.graphs[msg.Graph].UnmapInPort(msg.Public)
		UpdateComponentInfo(msg.Graph)
		entry, _ := ComponentRegistry[msg.Graph]
		sendJSON(ws, componentMessage{
			Protocol: "component",
			Command:  "component",
			Payload:  entry.Info,
		})
	}
	r.handlers["graph.renameinport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(renamePort)
		r.graphs[msg.Graph].RenameInPort(msg.From, msg.To)
		UpdateComponentInfo(msg.Graph)
		entry, _ := ComponentRegistry[msg.Graph]
		sendJSON(ws, componentMessage{
			Protocol: "component",
			Command:  "component",
			Payload:  entry.Info,
		})
	}
	r.handlers["graph.addoutport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(addPort)
		r.graphs[msg.Graph].MapOutPort(msg.Public, msg.Node, msg.Port)
		UpdateComponentInfo(msg.Graph)
		entry, _ := ComponentRegistry[msg.Graph]
		sendJSON(ws, componentMessage{
			Protocol: "component",
			Command:  "component",
			Payload:  entry.Info,
		})
	}
	r.handlers["graph.removeoutport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(removePort)
		r.graphs[msg.Graph].UnsetOutPort(msg.Public)
		r.graphs[msg.Graph].UnmapOutPort(msg.Public)
		UpdateComponentInfo(msg.Graph)
		entry, _ := ComponentRegistry[msg.Graph]
		sendJSON(ws, componentMessage{
			Protocol: "component",
			Command:  "component",
			Payload:  entry.Info,
		})
	}
	r.handlers["graph.renameoutport"] = func(ws *websocket.Conn, payload interface{}) {
		msg := payload.(renamePort)
		r.graphs[msg.Graph].RenameOutPort(msg.From, msg.To)
		UpdateComponentInfo(msg.Graph)
		entry, _ := ComponentRegistry[msg.Graph]
		sendJSON(ws, componentMessage{
			Protocol: "component",
			Command:  "component",
			Payload:  entry.Info,
		})
	}
	r.handlers["component.list"] = func(ws *websocket.Conn, payload interface{}) {
		for key, entry := range ComponentRegistry {
			if len(entry.Info.InPorts) == 0 && len(entry.Info.OutPorts) == 0 {
				// Need to obtain ports annotation for the first time
				UpdateComponentInfo(key)
			}
			sendJSON(ws, componentMessage{
				Protocol: "component",
				Command:  "component",
				Payload:  entry.Info,
			})
		}
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

func (r *Runtime) Handle(w http.ResponseWriter, req *http.Request) {
	ws, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println("Websocket upgrader failed", err)
		return
	}
	defer ws.Close()
	for {
		msgType, bytes, err := ws.ReadMessage()
		if err != nil {
			log.Println("Websocket read error:", err)
			break
		}
		if msgType != websocket.TextMessage {
			log.Println("Unexpected binary message")
			break
		}
		var msg Message
		err = json.Unmarshal(bytes, &msg)
		if err != nil {
			log.Println("JSON decoding error:", err)
			break
		}
		handler, exists := r.handlers[msg.Protocol+"."+msg.Command]
		if !exists {
			log.Printf("Unknown command: %s.%s\n", msg.Protocol, msg.Command)
			break
		}
		handler(ws, msg.Payload)
	}
}

func (r *Runtime) Listen(address string) {
	r.upgrader = websocket.Upgrader{}

	http.Handle("/", http.HandlerFunc(r.Handle))

	go func() {
		log.Fatal(http.ListenAndServe(address, nil))
	}()
	close(r.ready)

	// Wait for termination signal
	<-r.done
}
