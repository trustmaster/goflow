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

// portName stores full port name within the network.
type portName struct {
	proc string // Process name in the network
	port string // Port name of the process
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
	Wait *sync.WaitGroup
	// Net is a pointer to parent network.
	Net *Graph
	// procs contains the processes of the network.
	procs map[string]interface{}
	// inPorts maps network incoming ports to component ports.
	inPorts map[string]portName
	// outPorts maps network outgoing ports to component ports.
	outPorts map[string]portName
}

// InitGraphState method initializes graph fields and allocates memory.
func (n *Graph) InitGraphState() {
	n.Wait = new(sync.WaitGroup)
	n.procs = make(map[string]interface{}, DefaultNetworkCapacity)
	n.inPorts = make(map[string]portName, DefaultNetworkPortsNum)
	n.outPorts = make(map[string]portName, DefaultNetworkPortsNum)
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
func (n *Graph) AddNew(componentName string, processName string, initialPacket interface{}) bool {
	proc := Factory(componentName, initialPacket)
	return n.Add(proc, processName)
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
	if snet := sv.FieldByName("Graph"); snet.IsValid() {
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
	if stport.Kind() != reflect.Chan || stport.ChanDir()&reflect.SendDir == 0 {
		panic(senderName + "." + senderPort + " is not a valid output channel")
		return false
	}

	// Make a channel of an appropriate type
	chanType := reflect.ChanOf(reflect.BothDir, stport.Elem())
	channel := reflect.MakeChan(chanType, bufferSize)

	// Set the channel
	if sport.CanSet() {
		sport.Set(channel)
	} else {
		panic(senderName + "." + senderPort + " is not settable")
	}

	// Get the reciever port
	var rport reflect.Value

	// Check if receiver is a net
	if rnet := rv.FieldByName("Graph"); rnet.IsValid() {
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
		panic(receiverName + "." + receiverPort + " is not a valid output channel")
		return false
	}

	// Set the channel
	if rport.CanSet() {
		rport.Set(channel)
	} else {
		panic(receiverName + "." + receiverPort + " is not settable")
	}

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
	// We assume that pName contains a valid reference to a proc/port
	p := reflect.ValueOf(n.procs[pName.proc])
	var ret reflect.Value
	if p.Elem().FieldByName("Graph").IsValid() {
		// Is a subnet
		if sub, ok2 := p.Elem().FieldByName("Graph").Addr().Interface().(portMapper); ok2 {
			ret = sub.getInPort(pName.port)
		} else {
			panic("flow.Graph.getInPort(): Couldn't get portMapper")
		}
	} else {
		ret = p.Elem().FieldByName(pName.port)
	}
	return ret
}

// getOutPort returns the outport with given name as reflect.Value channel.
func (n *Graph) getOutPort(name string) reflect.Value {
	pName, ok := n.outPorts[name]
	if !ok {
		panic("flow.Graph.getOutPort(): Invalid outport name: " + name)
	}
	// We assume that pName contains a valid reference to a proc/port
	p := reflect.ValueOf(n.procs[pName.proc])
	var ret reflect.Value
	if p.Elem().FieldByName("Graph").IsValid() {
		// Is a subnet
		if sub, ok2 := p.Elem().FieldByName("Graph").Addr().Interface().(portMapper); ok2 {
			ret = sub.getOutPort(pName.port)
		} else {
			panic("flow.Graph.getOutPort(): Couldn't get portMapper")
		}
	} else {
		ret = p.Elem().FieldByName(pName.port)
	}
	return ret
}

// getWait returns net's wait group.
func (n *Graph) getWait() *sync.WaitGroup {
	return n.Wait
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
	if p, procFound := n.procs[procName]; procFound {
		if i, isNet := p.(portMapper); isNet {
			// Is a subnet
			ret = i.hasInPort(procPort)
		} else {
			// Is a proc
			f := reflect.ValueOf(p).Elem().FieldByName(procPort)
			ret = f.IsValid() && f.Kind() == reflect.Chan && (f.Type().ChanDir()&reflect.RecvDir) != 0
		}
		if !ret {
			panic("flow.Graph.MapInPort(): No such inport: " + procName + "." + procPort)
		}
	} else {
		panic("flow.Graph.MapInPort(): No such process: " + procName)
	}
	if ret {
		n.inPorts[name] = portName{proc: procName, port: procPort}
	}
	return ret
}

// MapOutPort adds an outport to the net and maps it to a contained proc's port.
// It returns true on success or panics and returns false on error.
func (n *Graph) MapOutPort(name, procName, procPort string) bool {
	ret := false
	// Check if target component and port exists
	if p, procFound := n.procs[procName]; procFound {
		if i, isNet := p.(portMapper); isNet {
			// Is a subnet
			ret = i.hasOutPort(procPort)
		} else {
			// Is a proc
			f := reflect.ValueOf(p).Elem().FieldByName(procPort)
			ret = f.IsValid() && f.Kind() == reflect.Chan && (f.Type().ChanDir()&reflect.SendDir) != 0
		}
		if !ret {
			panic("flow.Graph.MapOutPort(): No such outport: " + procName + "." + procPort)
		}
	} else {
		panic("flow.Graph.MapOutPort(): No such process: " + procName)
	}
	if ret {
		n.outPorts[name] = portName{proc: procName, port: procPort}
	}
	return ret
}

// run runs the network and waits for all processes to finish.
func (n *Graph) run() {
	// Add processes to the waitgroup before starting them
	n.Wait.Add(len(n.procs))
	for _, v := range n.procs {
		// Check if it is a net or proc
		r := reflect.ValueOf(v).Elem()
		if r.FieldByName("Graph").IsValid() {
			RunNet(v)
		} else {
			RunProc(v)
		}
	}
	// Wait for all processes to terminate
	n.Wait.Wait()
	// Check if there is a parent net
	if n.Net != nil {
		// Notify parent of finish
		n.Net.Wait.Done()
	}
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
	return res
}

// RunNet runs the network by starting all of its processes.
// It runs Init/Finish handlers if the network implements Initializable/Finalizable interfaces.
func RunNet(i interface{}) {
	// Call user init function if exists
	if initable, ok := i.(Initializable); ok {
		initable.Init()
	}

	// Run the contained processes
	go func() {
		// Should contain a pointer to graph as its field
		v := reflect.ValueOf(i).Elem()
		if v.Kind() == reflect.Struct {
			vGraph := v.FieldByName("Graph")
			if vGraph.IsValid() && vGraph.Type().Name() == "Graph" {
				if ctr, ok := vGraph.Addr().Interface().(netController); ok {
					ctr.run()
				}
			} else {
				panic("flow.RunNet(): argument is not a valid network subclass")
			}
		} else {
			panic("flow.RunNet(): argument is not a pointer to struct")
		}

		// Call user finish function if exists
		if finable, ok := i.(Finalizable); ok {
			finable.Finish()
		}
	}()
}
