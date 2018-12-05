package flow

import (
	"fmt"
	"reflect"
)

// iip stands for Initial Information Packet representation
// within the network.
type iip struct {
	data interface{}
	proc string // Target process name
	port string // Target port name
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

// sendIIPs sends Initial Information Packets upon network start
func (n *Graph) sendIIPs() error {
	// Send initial IPs
	for _, ip := range n.iips {
		// Get the reciever port channel
		var channel reflect.Value
		found := false

		// Try to find it among network inports
		for _, inPort := range n.inPorts {
			if inPort.proc == ip.proc && inPort.port == ip.port {
				channel = inPort.channel
				found = true
				break
			}
		}

		if !found {
			// Try to find among connections
			for _, conn := range n.connections {
				if conn.tgt.proc == ip.proc && conn.tgt.port == ip.port {
					channel = conn.channel
					found = true
					break
				}
			}
		}

		if !found {
			// Try to find a proc and attach a new channel to it
			recvPort, err := n.getProcPort(ip.proc, ip.port, reflect.RecvDir)
			if err != nil {
				return err
			}

			// Make a channel of an appropriate type
			chanType := reflect.ChanOf(reflect.BothDir, recvPort.Type().Elem())
			channel = reflect.MakeChan(chanType, n.conf.BufferSize)

			recvPort.Set(channel)
			found = true
		}

		if found {
			// Send data to the port
			channel.Send(reflect.ValueOf(ip.data))
		} else {
			return fmt.Errorf("IIP target not found: '%s.%s'"+ip.proc, ip.port)
		}
	}
	return nil
}
