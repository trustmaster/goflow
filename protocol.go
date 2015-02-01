package flow

// Message represents a single FBP protocol message
type Message struct {
	// Protocol is NoFlo protocol identifier:
	// "runtime", "component", "graph" or "network"
	Protocol string `json:"protocol"`
	// Command is a command to be executed within the protocol
	Command string `json:"command"`
	// Payload is JSON-encoded body of the message
	Payload interface{} `json:"payload"`
}

// runtimeInfo message contains response to runtime.getruntime request
type runtimeInfo struct {
	Type         string   `json:"type"`
	Version      string   `json:"version"`
	Capabilities []string `json:"capabilities"`
	Id           string   `json:"id"`
}

type runtimeMessage struct {
	Protocol string      `json:"protocol"`
	Command  string      `json:"command"`
	Payload  runtimeInfo `json:"payload"`
}

// clearGraph message is sent by client to create a new empty graph
type clearGraph struct {
	Id          string
	Name        string `json:",omitempty"` // ignored
	Library     string `json:",omitempty"` // ignored
	Main        bool   `json:",omitempty"`
	Icon        string `json:",omitempty"`
	Description string `json:",omitempty"`
}

// addNode message is sent by client to add a node to a graph
type addNode struct {
	Id        string
	Component string
	Graph     string
	Metadata  map[string]interface{} `json:",omitempty"` // ignored
}

// removeNode is a client message to remove a node from a graph
type removeNode struct {
	Id    string
	Graph string
}

// renameNode is a client message to rename a node in a graph
type renameNode struct {
	From  string
	To    string
	Graph string
}

// changeNode is a client message to change the metadata
// associated to a node in the graph
type changeNode struct { // ignored
	Id       string
	Graph    string
	Metadata map[string]interface{}
}

// addEdge is a client message to create a connection in a graph
type addEdge struct {
	Src struct {
		Node  string
		Port  string
		Index int `json:",omitempty"` // ignored
	}
	Tgt struct {
		Node  string
		Port  string
		Index int `json:",omitempty"` // ignored
	}
	Graph    string
	Metadata map[string]interface{} `json:",omitempty"` // ignored
}

// removeEdge is a client message to delete a connection from a graph
type removeEdge struct {
	Src struct {
		Node string
		Port string
	}
	Tgt struct {
		Node string
		Port string
	}
	Graph string
}

// changeEdge is a client message to change connection metadata
type changeEdge struct { // ignored
	Src struct {
		Node  string
		Port  string
		Index int `json:",omitempty"`
	}
	Tgt struct {
		Node  string
		Port  string
		Index int `json:",omitempty"`
	}
	Graph    string
	Metadata map[string]interface{}
}

// addInitial is a client message to add an IIP to a graph
type addInitial struct {
	Src struct {
		Data interface{}
	}
	Tgt struct {
		Node  string
		Port  string
		Index int `json:",omitempty"` // ignored
	}
	Graph    string
	Metadata map[string]interface{} `json:",omitempty"` // ignored
}

// removeInitial is a client message to remove an IIP from a graph
type removeInitial struct {
	Tgt struct {
		Node  string
		Port  string
		Index int `json:",omitempty"` // ignored
	}
	Graph string
}

// addPort is a client message to add an exported inport/outport to the graph
type addPort struct {
	Public   string
	Node     string
	Port     string
	Graph    string
	Metadata map[string]interface{} `json:",omitempty"` // ignored
}

// removePort is a client message to remove an exported inport/outport from the graph
type removePort struct {
	Public string
	Graph  string
}

// renamePort is a client message to rename a port of a graph
type renamePort struct {
	From  string
	To    string
	Graph string
}

// PortInfo represents a port to a runtime client
type PortInfo struct {
	Id          string        `json:"id"`
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Addressable bool          `json:"addressable"` // ignored
	Required    bool          `json:"required"`
	Values      []interface{} `json:"values"`  // ignored
	Default     interface{}   `json:"default"` // ignored
}

// ComponentInfo represents a component to a protocol client
type ComponentInfo struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Subgraph    bool       `json:"subgraph"`
	InPorts     []PortInfo `json:"inPorts"`
	OutPorts    []PortInfo `json:"outPorts"`
}

type componentMessage struct {
	Protocol string        `json:"protocol"`
	Command  string        `json:"command"`
	Payload  ComponentInfo `json:"payload"`
}
