package flow

import (
	"fmt"
	"reflect"
)

// getInPort returns the inport with given name as reflect.Value channel.
func (n *Graph) getInPort(name string) (reflect.Value, error) {
	pName, ok := n.inPorts[name]
	if !ok {
		return reflect.ValueOf(nil), fmt.Errorf("Inport not found: '%s'", name)
	}
	return pName.channel, nil
}

// getOutPort returns the outport with given name as reflect.Value channel.
func (n *Graph) getOutPort(name string) (reflect.Value, error) {
	pName, ok := n.outPorts[name]
	if !ok {
		return reflect.ValueOf(nil), fmt.Errorf("Outport not found: '%s'", name)
	}
	return pName.channel, nil
}

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
