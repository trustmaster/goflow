package flow

import (
	"reflect"
	"sync"
)

// Default channel buffer capacity
var DefaultBufferSize = 32

// Default capacity of network's processes/connections maps
var DefaultNetworkCapacity = 32

// Full port name within the network
type portName struct {
	proc string // Process name in the network
	port string // Port name of the process
}

// An interface used to obtain subnet's ports
type portMapper interface {
	getInPort(string) reflect.Value
	getOutPort(string) reflect.Value
	hasInPort(string) bool
	hasOutPort(string) bool
	SetInPort(string, interface{}) bool
	SetOutPort(string, interface{}) bool
}

// An interface used to run the subnet
type netController interface {
	getWait() *sync.WaitGroup
	run()
}

// Graph of processes connected with packet channels
type Graph struct {
	// Used for graceful network termination
	Wait *sync.WaitGroup
	// A pointer to parent network
	Net *Graph
	// Contains the processes
	procs map[string]interface{}
	// Maps network incoming ports to component ports
	inPorts map[string]portName
	// Maps network outgoing ports to component ports
	outPorts map[string]portName
}

// Initializes graph fields
func (n *Graph) InitGraphState() {
	n.Wait = new(sync.WaitGroup)
	n.procs = make(map[string]interface{}, DefaultNetworkCapacity)
	n.inPorts = make(map[string]portName, 16)
	n.outPorts = make(map[string]portName, 16)
}

// Adds a new process to the network
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

// Connects a sender to a receiver
func (n *Graph) Connect(senderName, senderPort, receiverName, receiverPort string, channel interface{}) bool {
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
	st := sv.Type()
	rp := reflect.ValueOf(receiver)
	rv := rp.Elem()
	rt := rv.Type()
	if !sv.CanSet() {
		panic(senderName + " is not settable")
		return false
	}
	if !rv.CanSet() {
		panic(receiverName + " is not settable")
		return false
	}

	// Make a channel of an appropriate type
	//chanType := sport.Type
	//channel := reflect.MakeChan(chanType, DefaultBufferSize)
	ch := reflect.ValueOf(channel)
	if !ch.IsValid() || ch.Kind() != reflect.Chan {
		panic("Passed channel is not valid")
		return false
	}

	// Get the actual ports and link them to the channel
	// Check if sender is a net
	if snet := sv.FieldByName("Graph"); snet.IsValid() {
		// Sender is a net
		if pm, isPm := snet.Addr().Interface().(portMapper); isPm {
			pm.SetOutPort(senderPort, channel)
		}
	} else {
		// Sender is a proc
		sport := sv.FieldByName(senderPort)
		stport, sfound := st.FieldByName(senderPort)
		// Ensure given ports are valid
		if !sfound {
			panic(senderName + "." + senderPort + " not found")
			return false
		}
		if stport.Type.Kind() != reflect.Chan || stport.Type.ChanDir()&reflect.SendDir == 0 {
			panic(senderName + "." + senderPort + " is not a valid output channel")
			return false
		}
		// Set the channel
		if sport.CanSet() {
			sport.Set(ch)
		} else {
			panic(senderName + "." + senderPort + " is not settable")
		}
	}

	// Check if receiver is a net
	if rnet := rv.FieldByName("Graph"); rnet.IsValid() {
		if pm, isPm := rnet.Addr().Interface().(portMapper); isPm {
			pm.SetInPort(receiverPort, channel)
		}
	} else {
		// Receiver is a proc
		rport := rv.FieldByName(receiverPort)
		rtport, rfound := rt.FieldByName(receiverPort)
		// Ensure given ports are valid
		if !rfound {
			panic(receiverName + "." + receiverPort + " not found")
			return false
		}
		if rtport.Type.Kind() != reflect.Chan || rtport.Type.ChanDir()&reflect.RecvDir == 0 {
			panic(receiverName + "." + receiverPort + " is not a valid output channel")
			return false
		}
		// Set the channel
		if rport.CanSet() {
			rport.Set(ch)
		} else {
			panic(receiverName + "." + receiverPort + " is not settable")
		}
	}

	return true
}

// Returns the inport with given name as reflect.Value channel
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

// Returns the outport with given name as reflect.Value channel
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

// Returns net's wait group
func (n *Graph) getWait() *sync.WaitGroup {
	return n.Wait
}

// Checks if the net has an inport with given name
func (n *Graph) hasInPort(name string) bool {
	_, has := n.inPorts[name]
	return has
}

// Checks if the net has an outport with given name
func (n *Graph) hasOutPort(name string) bool {
	_, has := n.outPorts[name]
	return has
}

// Adds an inport to the net and maps it to a contained proc's port
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

// Adds an outport to the net and maps it to a contained proc's port
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

// Runs the network and waits for all processes to finish
func (n *Graph) run() {
	hasParent := n.Net != nil
	if hasParent {
		// Notify parent of start
		n.Net.Wait.Add(1)
	}
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
	if hasParent {
		// Notify parent of finish
		n.Net.Wait.Done()
	}
}

// Assigns a channel to a network's inport to talk to the outer world
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

// Assigns a channel to a network's outport to talk to the outer world
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

// Runs the network by starting all of its processes.
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
