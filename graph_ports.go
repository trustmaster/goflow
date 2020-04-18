package goflow

import (
	"fmt"
	"reflect"
)

// port stores full port information within the network.
type port struct {
	// Address of the port in the graph
	addr address
	// Actual channel attached
	channel reflect.Value
	// Runtime info
	info PortInfo
}

// getOutPort returns the outport with given name as reflect.Value channel.
func (n *Graph) getOutPort(name string) (reflect.Value, error) {
	pName, ok := n.outPorts[name]
	if !ok {
		return reflect.ValueOf(nil), fmt.Errorf("outport not found: '%s'", name)
	}
	return pName.channel, nil
}

// MapInPort adds an inport to the net and maps it to a contained proc's port.
func (n *Graph) MapInPort(name, procName, procPort string) error {
	addr := parseAddress(procName, procPort)
	n.inPorts[name] = port{addr: addr}
	return nil
}

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

// MapOutPort adds an outport to the net and maps it to a contained proc's port.
func (n *Graph) MapOutPort(name, procName, procPort string) error {
	var channel reflect.Value
	var err error
	addr := parseAddress(procName, procPort)
	if p, procFound := n.procs[procName]; procFound {
		if g, isNet := p.(*Graph); isNet {
			// Is a subnet
			channel, err = g.getOutPort(procPort)
		} else {
			// Is a proc
			channel, err = n.getChan(addr, reflect.SendDir)
		}
	} else {
		return fmt.Errorf("could not map outport: process '%s' not found", procName)
	}
	if err == nil {
		n.outPorts[name] = port{addr: addr, channel: channel}
	}
	return err
}

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

// SetInPort assigns a channel to a network's inport to talk to the outer world.
func (n *Graph) SetInPort(name string, channel interface{}) error {
	p, ok := n.inPorts[name]
	if !ok {
		return fmt.Errorf("SetInPort: port '%s' not defined", name)
	}
	// Try to attach it
	port, err := n.getProcPort(p.addr.proc, p.addr.port, reflect.RecvDir)
	if err != nil {
		return fmt.Errorf("cannot set inport '%s': %v", name, err)
	}
	_, err = attachPort(port, p.addr, reflect.RecvDir, reflect.ValueOf(channel), n.conf.BufferSize)
	if err != nil {
		return fmt.Errorf("cannot attach inport '%s': %v", name, err)
	}
	// Save it in inPorts to be used with IIPs if needed
	p.channel = reflect.ValueOf(channel)
	n.inPorts[name] = p
	return nil
}

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

// SetOutPort assigns a channel to a network's outport to talk to the outer world.
// It returns true on success or false if the outport cannot be set.
func (n *Graph) SetOutPort(name string, channel interface{}) error {
	// Get the component's outport associated
	p, err := n.getOutPort(name)
	if err != nil {
		return err
	}
	// Try to set it
	if p.CanSet() {
		p.Set(reflect.ValueOf(channel))
	} else {
		return fmt.Errorf("Cannot set graph outport: '%s'", name)
	}
	// Save it in outPorts to be used later
	if p, ok := n.outPorts[name]; ok {
		p.channel = reflect.ValueOf(channel)
		n.outPorts[name] = p
	}
	return nil
}

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
