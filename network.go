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
	// iips contains initial IPs attached to the network
	iips []iip
	// done is used to let the outside world know when the net has finished its job
	done chan struct{}
	// ready is used to let the outside world know when the net is ready to accept input
	ready chan struct{}
}

// InitGraphState method initializes graph fields and allocates memory.
func (n *Graph) InitGraphState() {
	n.waitGrp = new(sync.WaitGroup)
	n.procs = make(map[string]interface{}, DefaultNetworkCapacity)
	n.inPorts = make(map[string]port, DefaultNetworkPortsNum)
	n.outPorts = make(map[string]port, DefaultNetworkPortsNum)
	n.connections = make([]connection, 0, DefaultNetworkCapacity)
	n.iips = make([]iip, 0, DefaultNetworkPortsNum)
	n.done = make(chan struct{})
	n.ready = make(chan struct{})
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

// AddNew creates a new process instance using component factory and adds it to the network.
func (n *Graph) AddNew(componentName string, processName string) bool {
	proc := Factory(componentName)
	return n.Add(proc, processName)
}

// AddIIP adds an Initial Information packet to the network
func (n *Graph) AddIIP(data interface{}, processName, portName string) bool {
	if _, exists := n.procs[processName]; exists {
		n.iips = append(n.iips, iip{data: data, proc: processName, port: portName})
		return true
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

	// Validate sender port
	stport := sport.Type()
	var channel reflect.Value
	if stport.Kind() == reflect.Slice {

		if sport.Type().Elem().Kind() == reflect.Chan && sport.Type().Elem().ChanDir()&reflect.SendDir != 0 {

			// Need to create a new channel and add it to the array
			chanType := reflect.ChanOf(reflect.BothDir, sport.Type().Elem().Elem())
			channel = reflect.MakeChan(chanType, bufferSize)
			sport.Set(reflect.Append(sport, channel))
		}
	} else if stport.Kind() == reflect.Chan && stport.ChanDir()&reflect.SendDir != 0 {
		// Make a channel of an appropriate type
		chanType := reflect.ChanOf(reflect.BothDir, stport.Elem())
		channel = reflect.MakeChan(chanType, bufferSize)
		// Set the channel
		if sport.CanSet() {
			sport.Set(channel)
		} else {
			panic(senderName + "." + senderPort + " is not settable")
		}
	}

	if channel.IsNil() {
		panic(senderName + "." + senderPort + " is not a valid output channel")
		return false
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

	// Set the channel
	if rport.CanSet() {
		rport.Set(channel)
	} else {
		panic(receiverName + "." + receiverPort + " is not settable")
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

// getOutPort returns the outport with given name as reflect.Value channel.
func (n *Graph) getOutPort(name string) reflect.Value {
	pName, ok := n.outPorts[name]
	if !ok {
		panic("flow.Graph.getOutPort(): Invalid outport name: " + name)
	}
	return pName.channel
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
	// Check if there is a parent net
	if n.Net != nil {
		// Notify parent of finish
		n.Net.waitGrp.Done()
	}
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
