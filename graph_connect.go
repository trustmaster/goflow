package goflow

import (
	"errors"
	"fmt"
	"reflect"
)

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

// getProcPort finds an assignable port field in one of the subprocesses
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
		return nilValue, fmt.Errorf("Connect error: '%s.%s' is not of the correct chan type", procName, portName)
	}

	// Check assignability
	if !portVal.CanSet() {
		return nilValue, fmt.Errorf("'%s.%s' is not assignable", procName, portName)
	}

	return portVal, nil
}

// findExistingChan returns a channel attached to receiver if it already exists among connections
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
