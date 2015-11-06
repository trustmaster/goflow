package flow

import (
	"reflect"
	"sync"
)

// DefaultBufferSize is the default channel buffer capacity.
var DefaultBufferSize = 0

// DefaultNetworkCapacity is the default capacity of network's processes/ports maps.
var DefaultNetworkCapacity = 32

// Default network output or input ports number
var DefaultNetworkPortsNum = 16

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

// portMapper interface is used to obtain subnet's ports.
type portMapper interface {
	getInPort(string) reflect.Value
	getOutPort(string) reflect.Value
	hasInPort(string) bool
	hasOutPort(string) bool
	listInPorts() map[string]port
	listOutPorts() map[string]port
	SetInPort(string, interface{}) bool
	SetOutPort(string, interface{}) bool
}

// netController interface is used to run a subnet.
type netController interface {
	getWait() *sync.WaitGroup
	run()
}

// Graph represents a graph of processes connected with packet channels.
type Graph struct {
	// Wait is used for graceful network termination.
	waitGrp *sync.WaitGroup
	// Net is a pointer to parent network.
	Net *Graph
	// procs contains the processes of the network.
	procs map[string]interface{}
	// inPorts maps network incoming ports to component ports.
	inPorts map[string]port
	// outPorts maps network outgoing ports to component ports.
	outPorts map[string]port
	// connections contains graph edges and channels.
	connections []connection
	// sendChanRefCount tracks how many sendports use the same channel
	sendChanRefCount map[uintptr]uint
	// sendChanMutex is used to synchronize operations on the sendChanRefCount map.
	sendChanMutex sync.Locker
	// iips contains initial IPs attached to the network
	iips []iip
	// done is used to let the outside world know when the net has finished its job
	done chan struct{}
	// ready is used to let the outside world know when the net is ready to accept input
	ready chan struct{}
	// isRunning indicates that the network is currently running
	isRunning bool
}

// InitGraphState method initializes graph fields and allocates memory.
func (n *Graph) InitGraphState() {
	n.waitGrp = new(sync.WaitGroup)
	n.procs = make(map[string]interface{}, DefaultNetworkCapacity)
	n.inPorts = make(map[string]port, DefaultNetworkPortsNum)
	n.outPorts = make(map[string]port, DefaultNetworkPortsNum)
	n.connections = make([]connection, 0, DefaultNetworkCapacity)
	n.sendChanRefCount = make(map[uintptr]uint, DefaultNetworkCapacity)
	n.sendChanMutex = new(sync.Mutex)
	n.iips = make([]iip, 0, DefaultNetworkPortsNum)
	n.done = make(chan struct{})
	n.ready = make(chan struct{})
}

// Canvas is a generic graph that is manipulated at run-time only
type Canvas struct {
	Graph
}

// NewGraph creates a new canvas graph that can be modified at run-time.
// Implements ComponentConstructor interace, so can it be used with Factory.
func NewGraph() interface{} {
	net := new(Canvas)
	net.InitGraphState()
	return net
}

// Register an empty graph component in the registry
func init() {
	Register("Graph", NewGraph)
	Annotate("Graph", ComponentInfo{
		Description: "A clear graph",
		Icon:        "cogs",
	})
}

// Increments SendChanRefCount
func (n *Graph) IncSendChanRefCount(c reflect.Value) {
	n.sendChanMutex.Lock()
	defer n.sendChanMutex.Unlock()

	ptr := c.Pointer()
	cnt := n.sendChanRefCount[ptr]
	cnt += 1
	n.sendChanRefCount[ptr] = cnt
}

// Decrements SendChanRefCount
// It returns true if the RefCount has reached 0
func (n *Graph) DecSendChanRefCount(c reflect.Value) bool {
	n.sendChanMutex.Lock()
	defer n.sendChanMutex.Unlock()

	ptr := c.Pointer()
	cnt := n.sendChanRefCount[ptr]
	if cnt == 0 {
		return true //yes you may try to close a nonexistant channel, see what happens...
	}
	cnt -= 1
	n.sendChanRefCount[ptr] = cnt
	return cnt == 0
}

// Add adds a new process with a given name to the network.
// It returns true on success or panics and returns false on error.
func (n *Graph) Add(c interface{}, name string) bool {
	// Check if passed interface is a valid pointer to struct
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		panic("flow.Graph.Add() argument is not a valid pointer")
		return false
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		panic("flow.Graph.Add() argument is not a valid pointer to struct")
		return false
	}
	// Set the link to self in the proccess so that it could use it
	var vNet reflect.Value
	vCom := v.FieldByName("Component")
	if vCom.IsValid() && vCom.Type().Name() == "Component" {
		vNet = vCom.FieldByName("Net")
	} else {
		vGraph := v.FieldByName("Graph")
		if vGraph.IsValid() && vGraph.Type().Name() == "Graph" {
			vNet = vGraph.FieldByName("Net")
		}
	}
	if vNet.IsValid() && vNet.CanSet() {
		vNet.Set(reflect.ValueOf(n))
	}
	// Add to the map of processes
	n.procs[name] = c
	return true
}

// AddGraph adds a new blank graph instance to a network. That instance can
// be modified then at run-time.
func (n *Graph) AddGraph(name string) bool {
	net := new(Graph)
	net.InitGraphState()
	return n.Add(net, name)
}

// AddNew creates a new process instance using component factory and adds it to the network.
func (n *Graph) AddNew(componentName string, processName string) bool {
	proc := Factory(componentName)
	return n.Add(proc, processName)
}

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

// Rename changes a process name in all connections, external ports, IIPs and the
// graph itself.
func (n *Graph) Rename(processName, newName string) bool {
	if _, exists := n.procs[processName]; !exists {
		return false
	}
	if _, busy := n.procs[newName]; busy {
		// New name is already taken
		return false
	}
	for i, conn := range n.connections {
		if conn.src.proc == processName {
			n.connections[i].src.proc = newName
		}
		if conn.tgt.proc == processName {
			n.connections[i].tgt.proc = newName
		}
	}
	for key, port := range n.inPorts {
		if port.proc == processName {
			tmp := n.inPorts[key]
			tmp.proc = newName
			n.inPorts[key] = tmp
		}
	}
	for key, port := range n.outPorts {
		if port.proc == processName {
			tmp := n.outPorts[key]
			tmp.proc = newName
			n.outPorts[key] = tmp
		}
	}
	n.procs[newName] = n.procs[processName]
	delete(n.procs, processName)
	return true
}

// AddIIP adds an Initial Information packet to the network
func (n *Graph) AddIIP(data interface{}, processName, portName string) bool {
	if _, exists := n.procs[processName]; exists {
		n.iips = append(n.iips, iip{data: data, proc: processName, port: portName})
		return true
	}
	return false
}

// RemoveIIP detaches an IIP from specific process and port
func (n *Graph) RemoveIIP(processName, portName string) bool {
	for i, p := range n.iips {
		if p.proc == processName && p.port == portName {
			// Remove item from the slice
			n.iips[len(n.iips)-1], n.iips[i], n.iips = iip{}, n.iips[len(n.iips)-1], n.iips[:len(n.iips)-1]
			return true
		}
	}
	return false
}

// Connect connects a sender to a receiver and creates a channel between them using DefaultBufferSize.
// Normally such a connection is unbuffered but you can change by setting flow.DefaultBufferSize > 0 or
// by using ConnectBuf() function instead.
// It returns true on success or panics and returns false if error occurs.
func (n *Graph) Connect(senderName, senderPort, receiverName, receiverPort string) bool {
	return n.ConnectBuf(senderName, senderPort, receiverName, receiverPort, DefaultBufferSize)
}

// Connect connects a sender to a receiver using a channel with a buffer of a given size.
// It returns true on success or panics and returns false if error occurs.
func (n *Graph) ConnectBuf(senderName, senderPort, receiverName, receiverPort string, bufferSize int) bool {
	// Ensure sender and receiver processes exist
	sender, senderFound := n.procs[senderName]
	receiver, receiverFound := n.procs[receiverName]
	if !senderFound {
		panic("Sender '" + senderName + "' not found")
		return false
	}
	if !receiverFound {
		panic("Receiver '" + receiverName + "' not found")
		return false
	}

	// Ensure sender and receiver are settable
	sp := reflect.ValueOf(sender)
	sv := sp.Elem()
	// st := sv.Type()
	rp := reflect.ValueOf(receiver)
	rv := rp.Elem()
	// rt := rv.Type()
	if !sv.CanSet() {
		panic(senderName + " is not settable")
		return false
	}
	if !rv.CanSet() {
		panic(receiverName + " is not settable")
		return false
	}

	var sport reflect.Value

	// Get the actual ports and link them to the channel
	// Check if sender is a net
	var snet reflect.Value
	if sv.Type().Name() == "Graph" {
		snet = sv
	} else {
		snet = sv.FieldByName("Graph")
	}
	if snet.IsValid() {
		// Sender is a net
		if pm, isPm := snet.Addr().Interface().(portMapper); isPm {
			sport = pm.getOutPort(senderPort)
		}
	} else {
		// Sender is a proc
		sport = sv.FieldByName(senderPort)
	}

	// Get the reciever port
	var rport reflect.Value

	// Check if receiver is a net
	var rnet reflect.Value
	if rv.Type().Name() == "Graph" {
		rnet = rv
	} else {
		rnet = rv.FieldByName("Graph")
	}
	if rnet.IsValid() {
		if pm, isPm := rnet.Addr().Interface().(portMapper); isPm {
			rport = pm.getInPort(receiverPort)
		}
	} else {
		// Receiver is a proc
		rport = rv.FieldByName(receiverPort)
	}

	// Validate receiver port
	rtport := rport.Type()
	if rtport.Kind() != reflect.Chan || rtport.ChanDir()&reflect.RecvDir == 0 {
		panic(receiverName + "." + receiverPort + " is not a valid input channel")
		return false
	}

	// Validate sender port
	stport := sport.Type()
	var channel reflect.Value
	if !rport.IsNil() {
		for _, mycon := range n.connections {
			if mycon.tgt.port == receiverPort && mycon.tgt.proc == receiverName {
				channel = mycon.channel
				break
			}
		}
	}
	if stport.Kind() == reflect.Slice {

		if sport.Type().Elem().Kind() == reflect.Chan && sport.Type().Elem().ChanDir()&reflect.SendDir != 0 {

			if !channel.IsValid() {
				// Need to create a new channel and add it to the array
				chanType := reflect.ChanOf(reflect.BothDir, sport.Type().Elem().Elem())
				channel = reflect.MakeChan(chanType, bufferSize)
			}
			sport.Set(reflect.Append(sport, channel))
			n.IncSendChanRefCount(channel)
		}
	} else if stport.Kind() == reflect.Chan && stport.ChanDir()&reflect.SendDir != 0 {
		// Check if channel was already instantiated, if so, use it. Thus we can connect serveral endpoints and golang will pseudo-randomly chooses a receiver
		// Also, this avoids crashes on <-net.Wait()
		if !sport.IsNil() {
			//go does not allow cast of unidir chan to bidir chan (for good reason)
			//but luckily we saved it, so we look it up
			if channel.IsValid() && sport.Addr() != rport.Addr() {
				panic("Trying to connect an already connected source to an already connected target")
			}
			for _, mycon := range n.connections {
				if mycon.src.port == senderPort && mycon.src.proc == senderName {
					channel = mycon.channel
					break
				}
			}
		}
		// either sport was nil or we did not find a previous channel instance
		if !channel.IsValid() {
			// Make a channel of an appropriate type
			chanType := reflect.ChanOf(reflect.BothDir, stport.Elem())
			channel = reflect.MakeChan(chanType, bufferSize)
		}
	}

	if channel.IsNil() {
		panic(senderName + "." + senderPort + " is not a valid output channel")
		return false
	}

	// Check if ?port.Set() would cause panic and if so ... panic
	if !sport.CanSet() {
		panic(senderName + "." + senderPort + " is not settable")
	}
	if !rport.CanSet() {
		panic(receiverName + "." + receiverPort + " is not settable")
	}
	// Set the channels
	if sport.IsNil() {
		//note that if sport is a slice, this does not run, instead see code above (== reflect.Slice)
		sport.Set(channel)
		n.IncSendChanRefCount(channel)
	}
	if rport.IsNil() {
		rport.Set(channel)
	}

	// Add connection info
	n.connections = append(n.connections, connection{
		src: portName{proc: senderName,
			port: senderPort},
		tgt: portName{proc: receiverName,
			port: receiverPort},
		channel: channel,
		buffer:  bufferSize})

	return true
}

// Unsets an port of a given process
func unsetProcPort(proc interface{}, portName string, isOut bool) bool {
	v := reflect.ValueOf(proc)
	var ch reflect.Value
	if v.Elem().FieldByName("Graph").IsValid() {
		if subnet, ok := v.Elem().FieldByName("Graph").Addr().Interface().(*Graph); ok {
			if isOut {
				ch = subnet.getOutPort(portName)
			} else {
				ch = subnet.getInPort(portName)
			}
		} else {
			return false
		}
	} else {
		ch = v.Elem().FieldByName(portName)
	}
	if !ch.IsValid() {
		return false
	}
	ch.Set(reflect.Zero(ch.Type()))
	return true
}

// Disconnect removes a connection between sender's outport and receiver's inport.
func (n *Graph) Disconnect(senderName, senderPort, receiverName, receiverPort string) bool {
	var sender, receiver interface{}
	var ok bool
	sender, ok = n.procs[senderName]
	if !ok {
		return false
	}
	receiver, ok = n.procs[receiverName]
	if !ok {
		return false
	}
	res := unsetProcPort(sender, senderPort, true)
	res = res && unsetProcPort(receiver, receiverPort, false)
	return res
}

// Get returns a node contained in the network by its name.
func (n *Graph) Get(processName string) interface{} {
	if proc, ok := n.procs[processName]; ok {
		return proc
	} else {
		panic("Process with name '" + processName + "' was not found")
	}
}

// getInPort returns the inport with given name as reflect.Value channel.
func (n *Graph) getInPort(name string) reflect.Value {
	pName, ok := n.inPorts[name]
	if !ok {
		panic("flow.Graph.getInPort(): Invalid inport name: " + name)
	}
	return pName.channel
}

// listInPorts returns information about graph inports and their types.
func (n *Graph) listInPorts() map[string]port {
	return n.inPorts
}

// getOutPort returns the outport with given name as reflect.Value channel.
func (n *Graph) getOutPort(name string) reflect.Value {
	pName, ok := n.outPorts[name]
	if !ok {
		panic("flow.Graph.getOutPort(): Invalid outport name: " + name)
	}
	return pName.channel
}

// listOutPorts returns information about graph outports and their types.
func (n *Graph) listOutPorts() map[string]port {
	return n.outPorts
}

// getWait returns net's wait group.
func (n *Graph) getWait() *sync.WaitGroup {
	return n.waitGrp
}

// hasInPort checks if the net has an inport with given name.
func (n *Graph) hasInPort(name string) bool {
	_, has := n.inPorts[name]
	return has
}

// hasOutPort checks if the net has an outport with given name.
func (n *Graph) hasOutPort(name string) bool {
	_, has := n.outPorts[name]
	return has
}

// MapInPort adds an inport to the net and maps it to a contained proc's port.
// It returns true on success or panics and returns false on error.
func (n *Graph) MapInPort(name, procName, procPort string) bool {
	ret := false
	// Check if target component and port exists
	var channel reflect.Value
	if p, procFound := n.procs[procName]; procFound {
		if i, isNet := p.(portMapper); isNet {
			// Is a subnet
			ret = i.hasInPort(procPort)
			channel = i.getInPort(procPort)
		} else {
			// Is a proc
			f := reflect.ValueOf(p).Elem().FieldByName(procPort)
			ret = f.IsValid() && f.Kind() == reflect.Chan && (f.Type().ChanDir()&reflect.RecvDir) != 0
			channel = f
		}
		if !ret {
			panic("flow.Graph.MapInPort(): No such inport: " + procName + "." + procPort)
		}
	} else {
		panic("flow.Graph.MapInPort(): No such process: " + procName)
	}
	if ret {
		n.inPorts[name] = port{proc: procName, port: procPort, channel: channel}
	}
	return ret
}

// AnnotateInPort sets optional run-time annotation for the port utilized by
// runtimes and FBP protocol clients.
func (n *Graph) AnnotateInPort(name string, info PortInfo) bool {
	port, exists := n.inPorts[name]
	if !exists {
		return false
	}
	port.info = info
	return true
}

// UnmapInPort removes an existing inport mapping
func (n *Graph) UnmapInPort(name string) bool {
	if _, exists := n.inPorts[name]; !exists {
		return false
	}
	delete(n.inPorts, name)
	return true
}

// MapOutPort adds an outport to the net and maps it to a contained proc's port.
// It returns true on success or panics and returns false on error.
func (n *Graph) MapOutPort(name, procName, procPort string) bool {
	ret := false
	// Check if target component and port exists
	var channel reflect.Value
	if p, procFound := n.procs[procName]; procFound {
		if i, isNet := p.(portMapper); isNet {
			// Is a subnet
			ret = i.hasOutPort(procPort)
			channel = i.getOutPort(procPort)
		} else {
			// Is a proc
			f := reflect.ValueOf(p).Elem().FieldByName(procPort)
			ret = f.IsValid() && f.Kind() == reflect.Chan && (f.Type().ChanDir()&reflect.SendDir) != 0
			channel = f
		}
		if !ret {
			panic("flow.Graph.MapOutPort(): No such outport: " + procName + "." + procPort)
		}
	} else {
		panic("flow.Graph.MapOutPort(): No such process: " + procName)
	}
	if ret {
		n.outPorts[name] = port{proc: procName, port: procPort, channel: channel}
	}
	return ret
}

// AnnotateOutPort sets optional run-time annotation for the port utilized by
// runtimes and FBP protocol clients.
func (n *Graph) AnnotateOutPort(name string, info PortInfo) bool {
	port, exists := n.outPorts[name]
	if !exists {
		return false
	}
	port.info = info
	return true
}

// UnmapOutPort removes an existing outport mapping
func (n *Graph) UnmapOutPort(name string) bool {
	if _, exists := n.outPorts[name]; !exists {
		return false
	}
	delete(n.outPorts, name)
	return true
}

// run runs the network and waits for all processes to finish.
func (n *Graph) run() {
	// Add processes to the waitgroup before starting them
	n.waitGrp.Add(len(n.procs))
	for _, v := range n.procs {
		// Check if it is a net or proc
		r := reflect.ValueOf(v).Elem()
		if r.FieldByName("Graph").IsValid() {
			RunNet(v)
		} else {
			RunProc(v)
		}
	}
	n.isRunning = true

	// Send initial IPs
	for _, ip := range n.iips {
		// Get the reciever port
		var rport reflect.Value
		found := false

		// Try to find it among network inports
		for _, inPort := range n.inPorts {
			if inPort.proc == ip.proc && inPort.port == ip.port {
				rport = inPort.channel
				found = true
				break
			}
		}

		if !found {
			// Try to find among connections
			for _, conn := range n.connections {
				if conn.tgt.proc == ip.proc && conn.tgt.port == ip.port {
					rport = conn.channel
					found = true
					break
				}
			}
		}

		if !found {
			// Try to find a proc and attach a new channel to it
			for procName, proc := range n.procs {
				if procName == ip.proc {
					// Check if receiver is a net
					rv := reflect.ValueOf(proc).Elem()
					var rnet reflect.Value
					if rv.Type().Name() == "Graph" {
						rnet = rv
					} else {
						rnet = rv.FieldByName("Graph")
					}
					if rnet.IsValid() {
						if pm, isPm := rnet.Addr().Interface().(portMapper); isPm {
							rport = pm.getInPort(ip.port)
						}
					} else {
						// Receiver is a proc
						rport = rv.FieldByName(ip.port)
					}

					// Validate receiver port
					rtport := rport.Type()
					if rtport.Kind() != reflect.Chan || rtport.ChanDir()&reflect.RecvDir == 0 {
						panic(ip.proc + "." + ip.port + " is not a valid input channel")
					}
					var channel reflect.Value

					// Make a channel of an appropriate type
					chanType := reflect.ChanOf(reflect.BothDir, rtport.Elem())
					channel = reflect.MakeChan(chanType, DefaultBufferSize)
					// Set the channel
					if rport.CanSet() {
						rport.Set(channel)
					} else {
						panic(ip.proc + "." + ip.port + " is not settable")
					}

					// Use the new channel to send the IIP
					rport = channel
					found = true
					break
				}
			}
		}

		if found {
			// Send data to the port
			rport.Send(reflect.ValueOf(ip.data))
		} else {
			panic("IIP target not found: " + ip.proc + "." + ip.port)
		}
	}

	// Let the outside world know that the network is ready
	close(n.ready)

	// Wait for all processes to terminate
	n.waitGrp.Wait()
	n.isRunning = false
	// Check if there is a parent net
	if n.Net != nil {
		// Notify parent of finish
		n.Net.waitGrp.Done()
	}
}

// RunProc starts a proc added to a net at run time
func (n *Graph) RunProc(procName string) bool {
	if !n.isRunning {
		return false
	}
	proc, ok := n.procs[procName]
	if !ok {
		return false
	}
	v := reflect.ValueOf(proc).Elem()
	n.waitGrp.Add(1)
	if v.FieldByName("Graph").IsValid() {
		RunNet(proc)
		return true
	} else {
		ok = RunProc(proc)
		if !ok {
			n.waitGrp.Done()
		}
		return ok
	}
}

// Stop terminates the network without closing any connections
func (n *Graph) Stop() {
	if !n.isRunning {
		return
	}
	for _, v := range n.procs {
		// Check if it is a net or proc
		r := reflect.ValueOf(v).Elem()
		if r.FieldByName("Graph").IsValid() {
			subnet, ok := r.FieldByName("Graph").Addr().Interface().(*Graph)
			if !ok {
				panic("Couldn't get graph interface")
			}
			subnet.Stop()
		} else {
			StopProc(v)
		}
	}
}

// StopProc stops a specific process in the net
func (n *Graph) StopProc(procName string) bool {
	if !n.isRunning {
		return false
	}
	proc, ok := n.procs[procName]
	if !ok {
		return false
	}
	v := reflect.ValueOf(proc).Elem()
	if v.FieldByName("Graph").IsValid() {
		subnet, ok := v.FieldByName("Graph").Addr().Interface().(*Graph)
		if !ok {
			panic("Couldn't get graph interface")
		}
		subnet.Stop()
	} else {
		return StopProc(proc)
	}
	return true
}

// Ready returns a channel that can be used to suspend the caller
// goroutine until the network is ready to accept input packets
func (n *Graph) Ready() <-chan struct{} {
	return n.ready
}

// Wait returns a channel that can be used to suspend the caller
// goroutine until the network finishes its job
func (n *Graph) Wait() <-chan struct{} {
	return n.done
}

// SetInPort assigns a channel to a network's inport to talk to the outer world.
// It returns true on success or false if the inport cannot be set.
func (n *Graph) SetInPort(name string, channel interface{}) bool {
	res := false
	// Get the component's inport associated
	p := n.getInPort(name)
	// Try to set it
	if p.CanSet() {
		p.Set(reflect.ValueOf(channel))
		res = true
	}
	// Save it in inPorts to be used with IIPs if needed
	if p, ok := n.inPorts[name]; ok {
		p.channel = reflect.ValueOf(channel)
		n.inPorts[name] = p
	}
	return res
}

// RenameInPort changes graph's inport name
func (n *Graph) RenameInPort(oldName, newName string) bool {
	if _, exists := n.inPorts[oldName]; !exists {
		return false
	}
	n.inPorts[newName] = n.inPorts[oldName]
	delete(n.inPorts, oldName)
	return true
}

// UnsetInPort removes an external inport from the graph
func (n *Graph) UnsetInPort(name string) bool {
	port, exists := n.inPorts[name]
	if !exists {
		return false
	}
	if proc, ok := n.procs[port.proc]; ok {
		unsetProcPort(proc, port.port, false)
	}
	delete(n.inPorts, name)
	return true
}

// SetOutPort assigns a channel to a network's outport to talk to the outer world.
// It returns true on success or false if the outport cannot be set.
func (n *Graph) SetOutPort(name string, channel interface{}) bool {
	res := false
	// Get the component's outport associated
	p := n.getOutPort(name)
	// Try to set it
	if p.CanSet() {
		p.Set(reflect.ValueOf(channel))
		res = true
	}
	// Save it in outPorts to be used later
	if p, ok := n.outPorts[name]; ok {
		p.channel = reflect.ValueOf(channel)
		n.outPorts[name] = p
	}
	return res
}

// RenameOutPort changes graph's outport name
func (n *Graph) RenameOutPort(oldName, newName string) bool {
	if _, exists := n.outPorts[oldName]; !exists {
		return false
	}
	n.outPorts[newName] = n.outPorts[oldName]
	delete(n.outPorts, oldName)
	return true
}

// UnsetOutPort removes an external outport from the graph
func (n *Graph) UnsetOutPort(name string) bool {
	port, exists := n.outPorts[name]
	if !exists {
		return false
	}
	if proc, ok := n.procs[port.proc]; ok {
		unsetProcPort(proc, port.proc, true)
	}
	delete(n.outPorts, name)
	return true
}

// RunNet runs the network by starting all of its processes.
// It runs Init/Finish handlers if the network implements Initializable/Finalizable interfaces.
func RunNet(i interface{}) {
	// Get the contained network
	net, isGraph := i.(*Graph)
	if !isGraph {
		v := reflect.ValueOf(i).Elem()
		if v.Kind() != reflect.Struct {
			panic("flow.RunNet(): argument is not a pointer to struct")
		}
		vGraph := v.FieldByName("Graph")
		if !vGraph.IsValid() || vGraph.Type().Name() != "Graph" {
			panic("flow.RunNet(): argument is not a valid graph instance")
		}
		net = vGraph.Addr().Interface().(*Graph)
	}

	// Call user init function if exists
	if initable, ok := i.(Initializable); ok {
		initable.Init()
	}

	// Run the contained processes
	go func() {
		net.run()

		// Call user finish function if exists
		if finable, ok := i.(Finalizable); ok {
			finable.Finish()
		}

		// Close the wait channel
		close(net.done)
	}()
}
