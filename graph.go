package flow

import (
	"fmt"
	"sync"
)

// GraphConfig sets up properties for a graph
type GraphConfig struct {
	Capacity   uint
	BufferSize int
}

// defaultGraphConfig provides defaults for GraphConfig
func defaultGraphConfig() GraphConfig {
	return GraphConfig{
		Capacity:   32,
		BufferSize: 0,
	}
}

// Graph represents a graph of processes connected with packet channels.
type Graph struct {
	// Configuration for the graph
	conf GraphConfig
	// Wait is used for graceful network termination.
	waitGrp *sync.WaitGroup
	// procs contains the processes of the network.
	procs map[string]interface{}
	// inPorts maps network incoming ports to component ports.
	inPorts map[string]port
	// outPorts maps network outgoing ports to component ports.
	outPorts map[string]port
	// connections contains graph edges and channels.
	connections []connection
	// sendChanRefCount tracks how many outports use the same channel
	sendChanRefCount map[uintptr]uint
	// sendChanMutex is used to synchronize operations on the sendChanRefCount map.
	sendChanMutex sync.Locker
	// iips contains initial IPs attached to the network
	iips []iip
}

// NewGraph returns a new initialized empty graph instance
func NewGraph(config ...GraphConfig) *Graph {
	conf := defaultGraphConfig()
	if len(config) == 1 {
		conf = config[0]
	}

	return &Graph{
		conf:             conf,
		waitGrp:          new(sync.WaitGroup),
		procs:            make(map[string]interface{}, conf.Capacity),
		inPorts:          make(map[string]port, conf.Capacity),
		outPorts:         make(map[string]port, conf.Capacity),
		connections:      make([]connection, 0, conf.Capacity),
		sendChanRefCount: make(map[uintptr]uint, conf.Capacity),
		sendChanMutex:    new(sync.Mutex),
		iips:             make([]iip, 0, conf.Capacity),
	}
}

// NewDefaultGraph is a ComponentConstructor for the factory
func NewDefaultGraph() interface{} {
	return NewGraph()
}

// // Register an empty graph component in the registry
// func init() {
// 	Register("Graph", NewDefaultGraph)
// 	Annotate("Graph", ComponentInfo{
// 		Description: "A clear graph",
// 		Icon:        "cogs",
// 	})
// }

// Add adds a new process with a given name to the network.
func (n *Graph) Add(name string, c interface{}) error {
	// c should be either graph or a component
	_, isComponent := c.(Component)
	_, isGraph := c.(Graph)
	if !isComponent && !isGraph {
		return fmt.Errorf("Could not add process '%s': instance is neither Component nor Graph", name)
	}
	// Add to the map of processes
	n.procs[name] = c
	return nil
}

// AddGraph adds a new blank graph instance to a network. That instance can
// be modified then at run-time.
func (n *Graph) AddGraph(name string) error {
	return n.Add(name, NewDefaultGraph())
}

// // AddNew creates a new process instance using component factory and adds it to the network.
// func (n *Graph) AddNew(processName string, componentName string) error {
// 	proc := Factory(componentName)
// 	return n.Add(processName, proc)
// }

// Remove deletes a process from the graph. First it stops the process if running.
// Then it disconnects it from other processes and removes the connections from
// the graph. Then it drops the process itself.
func (n *Graph) Remove(processName string) bool {
	if _, exists := n.procs[processName]; !exists {
		return false
	}
	// TODO disconnect before removal
	delete(n.procs, processName)
	return true
}

// // Rename changes a process name in all connections, external ports, IIPs and the
// // graph itself.
// func (n *Graph) Rename(processName, newName string) bool {
// 	if _, exists := n.procs[processName]; !exists {
// 		return false
// 	}
// 	if _, busy := n.procs[newName]; busy {
// 		// New name is already taken
// 		return false
// 	}
// 	for i, conn := range n.connections {
// 		if conn.src.proc == processName {
// 			n.connections[i].src.proc = newName
// 		}
// 		if conn.tgt.proc == processName {
// 			n.connections[i].tgt.proc = newName
// 		}
// 	}
// 	for key, port := range n.inPorts {
// 		if port.proc == processName {
// 			tmp := n.inPorts[key]
// 			tmp.proc = newName
// 			n.inPorts[key] = tmp
// 		}
// 	}
// 	for key, port := range n.outPorts {
// 		if port.proc == processName {
// 			tmp := n.outPorts[key]
// 			tmp.proc = newName
// 			n.outPorts[key] = tmp
// 		}
// 	}
// 	n.procs[newName] = n.procs[processName]
// 	delete(n.procs, processName)
// 	return true
// }

// // Get returns a node contained in the network by its name.
// func (n *Graph) Get(processName string) interface{} {
// 	if proc, ok := n.procs[processName]; ok {
// 		return proc
// 	} else {
// 		panic("Process with name '" + processName + "' was not found")
// 	}
// }

// // getWait returns net's wait group.
// func (n *Graph) getWait() *sync.WaitGroup {
// 	return n.waitGrp
// }

// Process runs the network
func (n *Graph) Process() {
	for _, i := range n.procs {
		c, ok := i.(Component)
		if !ok {
			continue
		}
		n.waitGrp.Add(1)
		w := Run(c)
		go func() {
			<-w
			n.waitGrp.Done()
		}()
	}
	n.waitGrp.Wait()
}
