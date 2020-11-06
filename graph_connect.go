package goflow

import (
	"fmt"
	"reflect"
)

// connection stores information about a connection within the net.
type connection struct {
	src     address
	tgt     address
	channel reflect.Value
	buffer  int
}

// Connect a sender to a receiver and create a channel between them using BufferSize graph configuration.
// Normally such a connection is unbuffered but you can change by setting flow.DefaultBufferSize > 0 or
// by using ConnectBuf() function instead.
func (n *Graph) Connect(senderName, senderPort, receiverName, receiverPort string) error {
	return n.ConnectBuf(senderName, senderPort, receiverName, receiverPort, n.conf.BufferSize)
}

// ConnectBuf connects a sender to a receiver using a channel with a buffer of a given size.
// It returns true on success or panics and returns false if error occurs.
func (n *Graph) ConnectBuf(senderName, senderPort, receiverName, receiverPort string, bufferSize int) error {
	sendAddr, err := parseAddress(senderName, senderPort)
	if err != nil {
		return fmt.Errorf("bad sender address: %w", err)
	}

	sendPort, err := n.getProcPort(senderName, sendAddr.port, reflect.SendDir)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	recvAddr, err := parseAddress(receiverName, receiverPort)
	if err != nil {
		return fmt.Errorf("bad receiver address: %w", err)
	}

	recvPort, err := n.getProcPort(receiverName, recvAddr.port, reflect.RecvDir)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	isNewChan := false // tells if a new channel will need to be created for this connection
	// Try to find an existing outbound channel from the same sender,
	// so it can be used as fan-out FIFO
	ch := n.findExistingChan(sendAddr, reflect.SendDir)
	if !ch.IsValid() || ch.IsNil() {
		// Then try to find an existing inbound channel to the same receiver,
		// so it can be used as a fan-in FIFO
		ch = n.findExistingChan(recvAddr, reflect.RecvDir)
		if ch.IsValid() && !ch.IsNil() {
			// Increase the number of listeners on this already used channel
			n.incChanListenersCount(ch)
		} else {
			isNewChan = true
		}
	}

	if ch, err = attachPort(sendPort, sendAddr, reflect.SendDir, ch, bufferSize); err != nil {
		return fmt.Errorf("connect '%s.%s': %w", senderName, senderPort, err)
	}

	if _, err = attachPort(recvPort, recvAddr, reflect.RecvDir, ch, bufferSize); err != nil {
		return fmt.Errorf("connect '%s.%s': %w", receiverName, receiverPort, err)
	}

	if isNewChan {
		// Register the first listener on a newly created channel
		n.incChanListenersCount(ch)
	}

	// Add connection info
	n.connections = append(n.connections, connection{
		src:     sendAddr,
		tgt:     recvAddr,
		channel: ch,
		buffer:  bufferSize,
	})

	return nil
}

// getProcPort finds an assignable port field in one of the subprocesses.
func (n *Graph) getProcPort(procName, portName string, dir reflect.ChanDir) (reflect.Value, error) {
	// Check if process exists
	proc, ok := n.procs[procName]
	if !ok {
		return reflect.Value{}, fmt.Errorf("getProcPort: process '%s' not found", procName)
	}

	// Check if process is settable
	val := reflect.ValueOf(proc)
	if val.Kind() == reflect.Ptr && val.IsValid() {
		val = val.Elem()
	}

	if !val.CanSet() {
		return reflect.Value{}, fmt.Errorf("getProcPort: process '%s' is not settable", procName)
	}

	// Get the port value
	var (
		portVal reflect.Value
		err     error
	)

	// Check if sender is a net
	net, ok := val.Interface().(Graph)
	if ok {
		// Sender is a net
		var ports map[string]port
		if dir == reflect.SendDir {
			ports = net.outPorts
		} else {
			ports = net.inPorts
		}

		p, ok := ports[portName]
		if !ok {
			return reflect.Value{}, fmt.Errorf("getProcPort: subgraph '%s' does not have inport '%s'", procName, portName)
		}

		portVal, err = net.getProcPort(p.addr.proc, p.addr.port, dir)
	} else {
		// Sender is a proc
		portVal = val.FieldByName(portName)
	}

	if err == nil && (!portVal.IsValid()) {
		err = fmt.Errorf("process '%s' does not have a valid port '%s'", procName, portName)
	}

	if err != nil {
		return reflect.Value{}, fmt.Errorf("getProcPort: %w", err)
	}

	return portVal, nil
}

func attachPort(port reflect.Value, addr address, dir reflect.ChanDir, ch reflect.Value, bufSize int) (reflect.Value, error) {
	switch addr.kind() {
	case portKindChan:
		return attachChanPort(port, dir, ch, bufSize)
	case portKindArray:
		return attachArrayPort(port, addr.index, dir, ch, bufSize)
	case portKindMap:
		return attachMapPort(port, addr.key, dir, ch, bufSize)
	case portKindNone: // makes go-lint happy
	}

	return reflect.Value{}, fmt.Errorf("invalid address %v", addr)
}

func attachChanPort(port reflect.Value, dir reflect.ChanDir, ch reflect.Value, bufSize int) (reflect.Value, error) {
	if err := validateChanDir(port.Type(), dir); err != nil {
		return ch, err
	}

	if !port.CanSet() {
		return ch, fmt.Errorf("port is not assignable")
	}

	ch = selectOrMakeChan(ch, port, port.Type().Elem(), bufSize)
	port.Set(ch)

	return ch, nil
}

func attachMapPort(port reflect.Value, key string, dir reflect.ChanDir, ch reflect.Value, bufSize int) (reflect.Value, error) {
	if err := validateChanDir(port.Type().Elem(), dir); err != nil {
		return ch, err
	}

	kv := reflect.ValueOf(key)
	item := port.MapIndex(kv)
	ch = selectOrMakeChan(ch, item, port.Type().Elem().Elem(), bufSize)

	if port.IsNil() {
		m := reflect.MakeMap(port.Type())
		port.Set(m)
	}

	port.SetMapIndex(kv, ch)

	return ch, nil
}

func attachArrayPort(port reflect.Value, key int, dir reflect.ChanDir, ch reflect.Value, bufSize int) (reflect.Value, error) {
	if err := validateChanDir(port.Type().Elem(), dir); err != nil {
		return ch, err
	}

	const scalingFactor = 2

	switch {
	case port.IsNil():
		// Allocate a new slice
		port.Set(reflect.MakeSlice(port.Type(), key+1, scalingFactor*(key+1)))
	case port.Cap() <= key:
		// Allocate a new slice and copy all of the old contents, fallthrough since length is always less than capacity.
		port.Set(reflect.AppendSlice(reflect.MakeSlice(port.Type(), 0, scalingFactor*(key+1)), port))
		fallthrough
	case port.Len() <= key:
		// Extend the slice
		port.SetLen(key + 1)
	}

	item := port.Index(key)
	ch = selectOrMakeChan(ch, item, port.Type().Elem().Elem(), bufSize)
	item.Set(ch)

	return ch, nil
}

func validateChanDir(portType reflect.Type, dir reflect.ChanDir) error {
	switch {
	case portType.Kind() != reflect.Chan:
		return fmt.Errorf("not a channel")
	case portType.ChanDir()&dir == 0:
		return fmt.Errorf("channel does not support direction %s", dir.String())
	}

	return nil
}

func selectOrMakeChan(new, existing reflect.Value, t reflect.Type, bufSize int) reflect.Value {
	switch {
	case new.IsValid() && !new.IsNil():
		return new
	case existing.IsValid() && !existing.IsNil():
		return existing
	}

	return reflect.MakeChan(reflect.ChanOf(reflect.BothDir, t), bufSize)
}

// findExistingChan returns a channel attached to receiver if it already exists among connections.
func (n *Graph) findExistingChan(addr address, dir reflect.ChanDir) reflect.Value {
	// Find existing channel attached to the receiver
	for i := range n.connections {
		if dir == reflect.SendDir && n.connections[i].src == addr ||
			dir == reflect.RecvDir && n.connections[i].tgt == addr {
			return n.connections[i].channel
		}
	}

	return reflect.Value{}
}

// incChanListenersCount increments SendChanRefCount.
// The count is needed when multiple senders are connected
// to the same receiver. When the network is terminated and
// senders need to close their output port, this counter
// can help to avoid closing the same channel multiple times.
func (n *Graph) incChanListenersCount(c reflect.Value) {
	n.chanListenersCountLock.Lock()
	defer n.chanListenersCountLock.Unlock()

	ptr := c.Pointer()
	cnt := n.chanListenersCount[ptr]
	cnt++
	n.chanListenersCount[ptr] = cnt
}

// decChanListenersCount decrements SendChanRefCount
// It returns true if the RefCount has reached 0.
func (n *Graph) decChanListenersCount(c reflect.Value) bool {
	n.chanListenersCountLock.Lock()
	defer n.chanListenersCountLock.Unlock()

	ptr := c.Pointer()
	cnt := n.chanListenersCount[ptr]

	if cnt == 0 {
		return true // yes you may try to close a nonexistent channel, see what happens...
	}

	cnt--
	n.chanListenersCount[ptr] = cnt

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
