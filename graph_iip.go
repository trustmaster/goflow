package goflow

import (
	"fmt"
	"reflect"
)

// iip stands for Initial Information Packet representation
// within the network.
type iip struct {
	data interface{}
	addr address
}

// AddIIP adds an Initial Information packet to the network
func (n *Graph) AddIIP(processName, portName string, data interface{}) error {
	addr := parseAddress(processName, portName)
	if _, exists := n.procs[processName]; exists {
		n.iips = append(n.iips, iip{data: data, addr: addr})
		return nil
	}
	return fmt.Errorf("AddIIP: could not find '%s'", addr)
}

// RemoveIIP detaches an IIP from specific process and port
func (n *Graph) RemoveIIP(processName, portName string) error {
	addr := parseAddress(processName, portName)
	for i := range n.iips {
		if n.iips[i].addr == addr {
			// Remove item from the slice
			n.iips[len(n.iips)-1], n.iips[i], n.iips = iip{}, n.iips[len(n.iips)-1], n.iips[:len(n.iips)-1]
			return nil
		}
	}
	return fmt.Errorf("RemoveIIP: could not find IIP for '%s'", addr)
}

// sendIIPs sends Initial Information Packets upon network start
func (n *Graph) sendIIPs() error {
	// Send initial IPs
	for i := range n.iips {
		ip := n.iips[i]

		// Get the receiver port channel
		var channel reflect.Value
		var found bool

		// Try to find it among network inports
		for j := range n.inPorts {
			if n.inPorts[j].addr == ip.addr {
				channel = n.inPorts[j].channel
				found = true
				break
			}
		}

		if !found {
			// Try to find among connections
			for j := range n.connections {
				if n.connections[j].tgt == ip.addr {
					channel = n.connections[j].channel
					found = true
					break
				}
			}
		}

		var shouldClose bool

		if !found {
			// Try to find a proc and attach a new channel to it
			recvPort, err := n.getProcPort(ip.addr.proc, ip.addr.port, reflect.RecvDir)
			if err != nil {
				return err
			}

			channel, err = attachPort(recvPort, ip.addr, reflect.RecvDir, reflect.ValueOf(nil), n.conf.BufferSize)
			if err != nil {
				return err
			}

			found = true
			shouldClose = true
		}

		if !found {
			return fmt.Errorf("IIP target not found: '%s'", ip.addr)
		}

		// Send data to the port
		go func(channel, data reflect.Value, close bool) {
			channel.Send(data)
			if close {
				channel.Close()
			}
		}(channel, reflect.ValueOf(ip.data), shouldClose)
	}

	return nil
}
