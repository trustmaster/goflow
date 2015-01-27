package flow

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

// runtimeInfo message contains response to runtime.getruntime request
type runtimeInfo struct {
	Type         string
	Version      string
	Capabilities []string
	Id           string
}

// clearGraph message is sent by client to create a new empty graph
type clearGraph struct {
	Id          string
	Name        string `json:",omitempty"` // ignored
	Library     string `json:",omitempty"` // ignored
	Main        bool   `json:",omitempty"`
	Icon        string `json:",omitempty"` // ignored
	Description string `json:",omitempty"` // ignored
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

// portInfo represents a port to a runtime client
type portInfo struct {
	Id          string
	Type        string
	Description string
	Addressable bool // ignored
	Required    bool
	Values      []interface{} // ignored
	Default     interface{}   // ignored
}

// componentInfo represents a component to a protocol client
type componentInfo struct {
	Name        string
	Description string
	Icon        string
	Subgraph    bool
	InPorts     []portInfo
	OutPorts    []portInfo
}
