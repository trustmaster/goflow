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

// MapInPort adds an inport to the net and maps it to a contained proc's port.
func (n *Graph) MapInPort(name, procName, procPort string) {
	addr := parseAddress(procName, procPort)
	n.inPorts[name] = port{addr: addr}
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
func (n *Graph) MapOutPort(name, procName, procPort string) {
	addr := parseAddress(procName, procPort)
	n.outPorts[name] = port{addr: addr}
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
	return n.setGraphPort(name, channel, reflect.RecvDir)
}

// SetOutPort assigns a channel to a network's outport to talk to the outer world.
// It returns true on success or false if the outport cannot be set.
func (n *Graph) SetOutPort(name string, channel interface{}) error {
	return n.setGraphPort(name, channel, reflect.SendDir)
}

func (n *Graph) setGraphPort(name string, channel interface{}, dir reflect.ChanDir) error {
	var ports map[string]port
	var dirDescr string
	if dir == reflect.SendDir {
		ports = n.outPorts
		dirDescr = "out"
	} else {
		ports = n.inPorts
		dirDescr = "in"
	}
	p, ok := ports[name]
	if !ok {
		return fmt.Errorf("setGraphPort: %s port '%s' not defined", dirDescr, name)
	}
	// Try to attach it
	port, err := n.getProcPort(p.addr.proc, p.addr.port, dir)
	if err != nil {
		return fmt.Errorf("setGraphPort: cannot set %s port '%s': %w", dirDescr, name, err)
	}
	_, err = attachPort(port, p.addr, dir, reflect.ValueOf(channel), n.conf.BufferSize)
	if err != nil {
		return fmt.Errorf("setGraphPort: cannot attach %s port '%s': %w", dirDescr, name, err)
	}
	// Save it in inPorts to be used with IIPs if needed
	p.channel = reflect.ValueOf(channel)
	ports[name] = p
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
