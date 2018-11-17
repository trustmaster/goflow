package flow

import (
	"errors"
	"fmt"
	"reflect"
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

// port stores full port information within the network.
type port struct {
	// Process name in the network
	proc string
	// Port name of the process
	port string
	// Actual channel attached
	channel reflect.Value
	// Runtime info
	info PortInfo
}

// portName stores full port name within the network.
type portName struct {
	// Process name in the network
	proc string
	// Port name of the process
	port string
}

// connection stores information about a connection within the net.
type connection struct {
	src     portName
	tgt     portName
	channel reflect.Value
	buffer  int
}

// iip stands for Initial Information Packet representation
// within the network.
type iip struct {
	data interface{}
	proc string // Target process name
	port string // Target port name
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

// IncSendChanRefCount increments SendChanRefCount.
// The count is needed when multiple senders are connected
// to the same receiver. When the network is terminated and
// senders need to close their output port, this counter
// can help to avoid closing the same channel multiple times.
func (n *Graph) incSendChanRefCount(c reflect.Value) {
	n.sendChanMutex.Lock()
	defer n.sendChanMutex.Unlock()

	ptr := c.Pointer()
	cnt := n.sendChanRefCount[ptr]
	cnt++
	n.sendChanRefCount[ptr] = cnt
}

// DecSendChanRefCount decrements SendChanRefCount
// It returns true if the RefCount has reached 0
func (n *Graph) decSendChanRefCount(c reflect.Value) bool {
	n.sendChanMutex.Lock()
	defer n.sendChanMutex.Unlock()

	ptr := c.Pointer()
	cnt := n.sendChanRefCount[ptr]
	if cnt == 0 {
		return true //yes you may try to close a nonexistant channel, see what happens...
	}
	cnt--
	n.sendChanRefCount[ptr] = cnt
	return cnt == 0
}

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

// // AddIIP adds an Initial Information packet to the network
// func (n *Graph) AddIIP(data interface{}, processName, portName string) bool {
// 	if _, exists := n.procs[processName]; exists {
// 		n.iips = append(n.iips, iip{data: data, proc: processName, port: portName})
// 		return true
// 	}
// 	return false
// }

// // RemoveIIP detaches an IIP from specific process and port
// func (n *Graph) RemoveIIP(processName, portName string) bool {
// 	for i, p := range n.iips {
// 		if p.proc == processName && p.port == portName {
// 			// Remove item from the slice
// 			n.iips[len(n.iips)-1], n.iips[i], n.iips = iip{}, n.iips[len(n.iips)-1], n.iips[:len(n.iips)-1]
// 			return true
// 		}
// 	}
// 	return false
// }

// Connect connects a sender to a receiver and creates a channel between them using BufferSize configuratio nof the graph.
// Normally such a connection is unbuffered but you can change by setting flow.DefaultBufferSize > 0 or
// by using ConnectBuf() function instead.
// It returns true on success or panics and returns false if error occurs.
func (n *Graph) Connect(senderName, senderPort, receiverName, receiverPort string) error {
	return n.ConnectBuf(senderName, senderPort, receiverName, receiverPort, n.conf.BufferSize)
}

// ConnectBuf connects a sender to a receiver using a channel with a buffer of a given size.
// It returns true on success or panics and returns false if error occurs.
func (n *Graph) ConnectBuf(senderName, senderPort, receiverName, receiverPort string, bufferSize int) error {
	senderPortVal, err := n.getProcPort(senderName, senderPort, reflect.SendDir)
	if err != nil {
		return err
	}

	receiverPortVal, err := n.getProcPort(receiverName, receiverPort, reflect.RecvDir)
	if err != nil {
		return err
	}

	// Try to get an existing channel
	var channel reflect.Value
	if !receiverPortVal.IsNil() {
		// Find existing channel attached to the receiver
		channel = n.findExistingChan(receiverName, receiverPort, reflect.RecvDir)
	}

	sndPortType := senderPortVal.Type()

	if !senderPortVal.IsNil() {
		// If both ports are already busy, we cannot connect them
		if channel.IsValid() && senderPortVal.Addr() != receiverPortVal.Addr() {
			return fmt.Errorf("'%s.%s' cannot be connected to '%s.%s': both ports already in use", receiverName, receiverPort, senderName, senderPort)
		}
		// Find an existing channel attached to sender
		// Receiver channel takes priority if exists
		if !channel.IsValid() {
			channel = n.findExistingChan(senderName, senderPort, reflect.SendDir)
		}
	}

	// Create a new channel if none of the existing channles found
	if !channel.IsValid() {
		// Make a channel of an appropriate type
		chanType := reflect.ChanOf(reflect.BothDir, sndPortType.Elem())
		channel = reflect.MakeChan(chanType, bufferSize)
	}

	// Set the channels
	// TODO fix rewiring a graph without disconnecting ports
	if senderPortVal.IsNil() {
		senderPortVal.Set(channel)
		n.incSendChanRefCount(channel)
	}
	if receiverPortVal.IsNil() {
		receiverPortVal.Set(channel)
	}

	// Add connection info
	n.connections = append(n.connections, connection{
		src: portName{proc: senderName,
			port: senderPort},
		tgt: portName{proc: receiverName,
			port: receiverPort},
		channel: channel,
		buffer:  bufferSize})

	return nil
}

func (n *Graph) getProcPort(procName, portName string, dir reflect.ChanDir) (reflect.Value, error) {
	nilValue := reflect.ValueOf(nil)
	// Ensure process exists
	proc, ok := n.procs[procName]
	if !ok {
		return nilValue, fmt.Errorf("Connect error: process '%s' not found", procName)
	}

	// Ensure sender is settable
	val := reflect.ValueOf(proc)
	if val.Kind() == reflect.Ptr && val.IsValid() {
		val = val.Elem()
	}
	if !val.CanSet() {
		return nilValue, fmt.Errorf("Connect error: process '%s' is not settable", procName)
	}

	// Get the port value
	var portVal reflect.Value
	var err error
	// Check if sender is a net
	net, ok := val.Interface().(Graph)
	if ok {
		// Sender is a net
		if dir == reflect.SendDir {
			portVal, err = net.getOutPort(portName)
		} else {
			portVal, err = net.getInPort(portName)
		}

	} else {
		// Sender is a proc
		portVal = val.FieldByName(portName)
		if !portVal.IsValid() {
			err = errors.New("")
		}
	}
	if err != nil {
		return nilValue, fmt.Errorf("Connect error: process '%s' does not have port '%s'", procName, portName)
	}

	// Validate port type
	portType := portVal.Type()

	// Sender port can be an array port
	if dir == reflect.SendDir && portType.Kind() == reflect.Slice {
		portType = portType.Elem()
	}

	// Validate
	if portType.Kind() != reflect.Chan || portType.ChanDir()&dir == 0 {
		return nilValue, fmt.Errorf("Connect error: '%s.%s' does not have a valid chan type", procName, portName)
	}

	// Check assignability
	if !portVal.CanSet() {
		return nilValue, fmt.Errorf("'%s.%s' is not assignable", procName, portName)
	}

	return portVal, nil
}

// findExistingChan returns a channel attached to receiver if it already exists
func (n *Graph) findExistingChan(proc, procPort string, dir reflect.ChanDir) reflect.Value {
	var channel reflect.Value
	// Find existing channel attached to the receiver
	for _, conn := range n.connections {
		var p portName
		if dir == reflect.SendDir {
			p = conn.src
		} else {
			p = conn.tgt
		}
		if p.port == procPort && p.proc == proc {
			channel = conn.channel
			break
		}
	}
	return channel
}

// // Unsets an port of a given process
// func unsetProcPort(proc interface{}, portName string, isOut bool) bool {
// 	v := reflect.ValueOf(proc)
// 	var ch reflect.Value
// 	if v.Elem().FieldByName("Graph").IsValid() {
// 		if subnet, ok := v.Elem().FieldByName("Graph").Addr().Interface().(*Graph); ok {
// 			if isOut {
// 				ch = subnet.getOutPort(portName)
// 			} else {
// 				ch = subnet.getInPort(portName)
// 			}
// 		} else {
// 			return false
// 		}
// 	} else {
// 		ch = v.Elem().FieldByName(portName)
// 	}
// 	if !ch.IsValid() {
// 		return false
// 	}
// 	ch.Set(reflect.Zero(ch.Type()))
// 	return true
// }

// // Disconnect removes a connection between sender's outport and receiver's inport.
// func (n *Graph) Disconnect(senderName, senderPort, receiverName, receiverPort string) bool {
// 	var sender, receiver interface{}
// 	var ok bool
// 	sender, ok = n.procs[senderName]
// 	if !ok {
// 		return false
// 	}
// 	receiver, ok = n.procs[receiverName]
// 	if !ok {
// 		return false
// 	}
// 	res := unsetProcPort(sender, senderPort, true)
// 	res = res && unsetProcPort(receiver, receiverPort, false)
// 	return res
// }

// // Get returns a node contained in the network by its name.
// func (n *Graph) Get(processName string) interface{} {
// 	if proc, ok := n.procs[processName]; ok {
// 		return proc
// 	} else {
// 		panic("Process with name '" + processName + "' was not found")
// 	}
// }

// getInPort returns the inport with given name as reflect.Value channel.
func (n *Graph) getInPort(name string) (reflect.Value, error) {
	pName, ok := n.inPorts[name]
	if !ok {
		return reflect.ValueOf(nil), fmt.Errorf("Inport not found: '%s'", name)
	}
	return pName.channel, nil
}

// // listInPorts returns information about graph inports and their types.
// func (n *Graph) listInPorts() map[string]port {
// 	return n.inPorts
// }

// getOutPort returns the outport with given name as reflect.Value channel.
func (n *Graph) getOutPort(name string) (reflect.Value, error) {
	pName, ok := n.outPorts[name]
	if !ok {
		return reflect.ValueOf(nil), fmt.Errorf("Outport not found: '%s'", name)
	}
	return pName.channel, nil
}

// // listOutPorts returns information about graph outports and their types.
// func (n *Graph) listOutPorts() map[string]port {
// 	return n.outPorts
// }

// // getWait returns net's wait group.
// func (n *Graph) getWait() *sync.WaitGroup {
// 	return n.waitGrp
// }

// // hasInPort checks if the net has an inport with given name.
// func (n *Graph) hasInPort(name string) bool {
// 	_, has := n.inPorts[name]
// 	return has
// }

// // hasOutPort checks if the net has an outport with given name.
// func (n *Graph) hasOutPort(name string) bool {
// 	_, has := n.outPorts[name]
// 	return has
// }

// // MapInPort adds an inport to the net and maps it to a contained proc's port.
// // It returns true on success or panics and returns false on error.
// func (n *Graph) MapInPort(name, procName, procPort string) bool {
// 	ret := false
// 	// Check if target component and port exists
// 	var channel reflect.Value
// 	if p, procFound := n.procs[procName]; procFound {
// 		if i, isNet := p.(portMapper); isNet {
// 			// Is a subnet
// 			ret = i.hasInPort(procPort)
// 			channel = i.getInPort(procPort)
// 		} else {
// 			// Is a proc
// 			f := reflect.ValueOf(p).Elem().FieldByName(procPort)
// 			ret = f.IsValid() && f.Kind() == reflect.Chan && (f.Type().ChanDir()&reflect.RecvDir) != 0
// 			channel = f
// 		}
// 		if !ret {
// 			panic("flow.Graph.MapInPort(): No such inport: " + procName + "." + procPort)
// 		}
// 	} else {
// 		panic("flow.Graph.MapInPort(): No such process: " + procName)
// 	}
// 	if ret {
// 		n.inPorts[name] = port{proc: procName, port: procPort, channel: channel}
// 	}
// 	return ret
// }

// // AnnotateInPort sets optional run-time annotation for the port utilized by
// // runtimes and FBP protocol clients.
// func (n *Graph) AnnotateInPort(name string, info PortInfo) bool {
// 	port, exists := n.inPorts[name]
// 	if !exists {
// 		return false
// 	}
// 	port.info = info
// 	return true
// }

// // UnmapInPort removes an existing inport mapping
// func (n *Graph) UnmapInPort(name string) bool {
// 	if _, exists := n.inPorts[name]; !exists {
// 		return false
// 	}
// 	delete(n.inPorts, name)
// 	return true
// }

// // MapOutPort adds an outport to the net and maps it to a contained proc's port.
// // It returns true on success or panics and returns false on error.
// func (n *Graph) MapOutPort(name, procName, procPort string) bool {
// 	ret := false
// 	// Check if target component and port exists
// 	var channel reflect.Value
// 	if p, procFound := n.procs[procName]; procFound {
// 		if i, isNet := p.(portMapper); isNet {
// 			// Is a subnet
// 			ret = i.hasOutPort(procPort)
// 			channel = i.getOutPort(procPort)
// 		} else {
// 			// Is a proc
// 			f := reflect.ValueOf(p).Elem().FieldByName(procPort)
// 			ret = f.IsValid() && f.Kind() == reflect.Chan && (f.Type().ChanDir()&reflect.SendDir) != 0
// 			channel = f
// 		}
// 		if !ret {
// 			panic("flow.Graph.MapOutPort(): No such outport: " + procName + "." + procPort)
// 		}
// 	} else {
// 		panic("flow.Graph.MapOutPort(): No such process: " + procName)
// 	}
// 	if ret {
// 		n.outPorts[name] = port{proc: procName, port: procPort, channel: channel}
// 	}
// 	return ret
// }

// // AnnotateOutPort sets optional run-time annotation for the port utilized by
// // runtimes and FBP protocol clients.
// func (n *Graph) AnnotateOutPort(name string, info PortInfo) bool {
// 	port, exists := n.outPorts[name]
// 	if !exists {
// 		return false
// 	}
// 	port.info = info
// 	return true
// }

// // UnmapOutPort removes an existing outport mapping
// func (n *Graph) UnmapOutPort(name string) bool {
// 	if _, exists := n.outPorts[name]; !exists {
// 		return false
// 	}
// 	delete(n.outPorts, name)
// 	return true
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

// // SetInPort assigns a channel to a network's inport to talk to the outer world.
// // It returns true on success or false if the inport cannot be set.
// func (n *Graph) SetInPort(name string, channel interface{}) bool {
// 	res := false
// 	// Get the component's inport associated
// 	p := n.getInPort(name)
// 	// Try to set it
// 	if p.CanSet() {
// 		p.Set(reflect.ValueOf(channel))
// 		res = true
// 	}
// 	// Save it in inPorts to be used with IIPs if needed
// 	if p, ok := n.inPorts[name]; ok {
// 		p.channel = reflect.ValueOf(channel)
// 		n.inPorts[name] = p
// 	}
// 	return res
// }

// // RenameInPort changes graph's inport name
// func (n *Graph) RenameInPort(oldName, newName string) bool {
// 	if _, exists := n.inPorts[oldName]; !exists {
// 		return false
// 	}
// 	n.inPorts[newName] = n.inPorts[oldName]
// 	delete(n.inPorts, oldName)
// 	return true
// }

// // UnsetInPort removes an external inport from the graph
// func (n *Graph) UnsetInPort(name string) bool {
// 	port, exists := n.inPorts[name]
// 	if !exists {
// 		return false
// 	}
// 	if proc, ok := n.procs[port.proc]; ok {
// 		unsetProcPort(proc, port.port, false)
// 	}
// 	delete(n.inPorts, name)
// 	return true
// }

// // SetOutPort assigns a channel to a network's outport to talk to the outer world.
// // It returns true on success or false if the outport cannot be set.
// func (n *Graph) SetOutPort(name string, channel interface{}) bool {
// 	res := false
// 	// Get the component's outport associated
// 	p := n.getOutPort(name)
// 	// Try to set it
// 	if p.CanSet() {
// 		p.Set(reflect.ValueOf(channel))
// 		res = true
// 	}
// 	// Save it in outPorts to be used later
// 	if p, ok := n.outPorts[name]; ok {
// 		p.channel = reflect.ValueOf(channel)
// 		n.outPorts[name] = p
// 	}
// 	return res
// }

// // RenameOutPort changes graph's outport name
// func (n *Graph) RenameOutPort(oldName, newName string) bool {
// 	if _, exists := n.outPorts[oldName]; !exists {
// 		return false
// 	}
// 	n.outPorts[newName] = n.outPorts[oldName]
// 	delete(n.outPorts, oldName)
// 	return true
// }

// // UnsetOutPort removes an external outport from the graph
// func (n *Graph) UnsetOutPort(name string) bool {
// 	port, exists := n.outPorts[name]
// 	if !exists {
// 		return false
// 	}
// 	if proc, ok := n.procs[port.proc]; ok {
// 		unsetProcPort(proc, port.proc, true)
// 	}
// 	delete(n.outPorts, name)
// 	return true
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
