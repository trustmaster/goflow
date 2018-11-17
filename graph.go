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
	// done is used to let the outside world know when the net has finished its job
	done Wait
	// ready is used to let the outside world know when the net is ready to accept input
	ready Wait
	// isRunning indicates that the network is currently running
	isRunning bool
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
		done:             make(Wait),
		ready:            make(Wait),
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

// // run runs the network and waits for all processes to finish.
// func (n *Graph) run() {
// 	// Add processes to the waitgroup before starting them
// 	n.waitGrp.Add(len(n.procs))
// 	for _, v := range n.procs {
// 		// Check if it is a net or proc
// 		r := reflect.ValueOf(v).Elem()
// 		if r.FieldByName("Graph").IsValid() {
// 			RunNet(v)
// 		} else {
// 			RunProc(v)
// 		}
// 	}
// 	n.isRunning = true

// 	// Send initial IPs
// 	for _, ip := range n.iips {
// 		// Get the reciever port
// 		var rport reflect.Value
// 		found := false

// 		// Try to find it among network inports
// 		for _, inPort := range n.inPorts {
// 			if inPort.proc == ip.proc && inPort.port == ip.port {
// 				rport = inPort.channel
// 				found = true
// 				break
// 			}
// 		}

// 		if !found {
// 			// Try to find among connections
// 			for _, conn := range n.connections {
// 				if conn.tgt.proc == ip.proc && conn.tgt.port == ip.port {
// 					rport = conn.channel
// 					found = true
// 					break
// 				}
// 			}
// 		}

// 		if !found {
// 			// Try to find a proc and attach a new channel to it
// 			for procName, proc := range n.procs {
// 				if procName == ip.proc {
// 					// Check if receiver is a net
// 					rv := reflect.ValueOf(proc).Elem()
// 					var rnet reflect.Value
// 					if rv.Type().Name() == "Graph" {
// 						rnet = rv
// 					} else {
// 						rnet = rv.FieldByName("Graph")
// 					}
// 					if rnet.IsValid() {
// 						if pm, isPm := rnet.Addr().Interface().(portMapper); isPm {
// 							rport = pm.getInPort(ip.port)
// 						}
// 					} else {
// 						// Receiver is a proc
// 						rport = rv.FieldByName(ip.port)
// 					}

// 					// Validate receiver port
// 					rtport := rport.Type()
// 					if rtport.Kind() != reflect.Chan || rtport.ChanDir()&reflect.RecvDir == 0 {
// 						panic(ip.proc + "." + ip.port + " is not a valid input channel")
// 					}
// 					var channel reflect.Value

// 					// Make a channel of an appropriate type
// 					chanType := reflect.ChanOf(reflect.BothDir, rtport.Elem())
// 					channel = reflect.MakeChan(chanType, DefaultBufferSize)
// 					// Set the channel
// 					if rport.CanSet() {
// 						rport.Set(channel)
// 					} else {
// 						panic(ip.proc + "." + ip.port + " is not settable")
// 					}

// 					// Use the new channel to send the IIP
// 					rport = channel
// 					found = true
// 					break
// 				}
// 			}
// 		}

// 		if found {
// 			// Send data to the port
// 			rport.Send(reflect.ValueOf(ip.data))
// 		} else {
// 			panic("IIP target not found: " + ip.proc + "." + ip.port)
// 		}
// 	}

// 	// Let the outside world know that the network is ready
// 	close(n.ready)

// 	// Wait for all processes to terminate
// 	n.waitGrp.Wait()
// 	n.isRunning = false
// 	// Check if there is a parent net
// 	if n.Net != nil {
// 		// Notify parent of finish
// 		n.Net.waitGrp.Done()
// 	}
// }

// // RunProc starts a proc added to a net at run time
// func (n *Graph) RunProc(procName string) bool {
// 	if !n.isRunning {
// 		return false
// 	}
// 	proc, ok := n.procs[procName]
// 	if !ok {
// 		return false
// 	}
// 	v := reflect.ValueOf(proc).Elem()
// 	n.waitGrp.Add(1)
// 	if v.FieldByName("Graph").IsValid() {
// 		RunNet(proc)
// 		return true
// 	} else {
// 		ok = RunProc(proc)
// 		if !ok {
// 			n.waitGrp.Done()
// 		}
// 		return ok
// 	}
// }

// // Stop terminates the network without closing any connections
// func (n *Graph) Stop() {
// 	if !n.isRunning {
// 		return
// 	}
// 	for _, v := range n.procs {
// 		// Check if it is a net or proc
// 		r := reflect.ValueOf(v).Elem()
// 		if r.FieldByName("Graph").IsValid() {
// 			subnet, ok := r.FieldByName("Graph").Addr().Interface().(*Graph)
// 			if !ok {
// 				panic("Couldn't get graph interface")
// 			}
// 			subnet.Stop()
// 		} else {
// 			StopProc(v)
// 		}
// 	}
// }

// // StopProc stops a specific process in the net
// func (n *Graph) StopProc(procName string) bool {
// 	if !n.isRunning {
// 		return false
// 	}
// 	proc, ok := n.procs[procName]
// 	if !ok {
// 		return false
// 	}
// 	v := reflect.ValueOf(proc).Elem()
// 	if v.FieldByName("Graph").IsValid() {
// 		subnet, ok := v.FieldByName("Graph").Addr().Interface().(*Graph)
// 		if !ok {
// 			panic("Couldn't get graph interface")
// 		}
// 		subnet.Stop()
// 	} else {
// 		return StopProc(proc)
// 	}
// 	return true
// }

// // Ready returns a channel that can be used to suspend the caller
// // goroutine until the network is ready to accept input packets
// func (n *Graph) Ready() <-chan struct{} {
// 	return n.ready
// }

// // Wait returns a channel that can be used to suspend the caller
// // goroutine until the network finishes its job
// func (n *Graph) Wait() <-chan struct{} {
// 	return n.done
// }

// // RunNet runs the network by starting all of its processes.
// // It runs Init/Finish handlers if the network implements Initializable/Finalizable interfaces.
// func RunNet(i interface{}) {
// 	// Get the contained network
// 	net, isGraph := i.(*Graph)
// 	if !isGraph {
// 		v := reflect.ValueOf(i).Elem()
// 		if v.Kind() != reflect.Struct {
// 			panic("flow.RunNet(): argument is not a pointer to struct")
// 		}
// 		vGraph := v.FieldByName("Graph")
// 		if !vGraph.IsValid() || vGraph.Type().Name() != "Graph" {
// 			panic("flow.RunNet(): argument is not a valid graph instance")
// 		}
// 		net = vGraph.Addr().Interface().(*Graph)
// 	}

// 	// Call user init function if exists
// 	if initable, ok := i.(Initializable); ok {
// 		initable.Init()
// 	}

// 	// Run the contained processes
// 	go func() {
// 		net.run()

// 		// Call user finish function if exists
// 		if finable, ok := i.(Finalizable); ok {
// 			finable.Finish()
// 		}

// 		// Close the wait channel
// 		close(net.done)
// 	}()
// }
