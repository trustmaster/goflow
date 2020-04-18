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
	for i, p := range n.iips {
		if p.addr == addr {
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
	for _, ip := range n.iips {
		ip := ip
		// Get the reciever port channel
		var channel reflect.Value
		found := false
		shouldClose := false

		// Try to find it among network inports
		for _, inPort := range n.inPorts {
			if inPort.addr == ip.addr {
				channel = inPort.channel
				found = true
				break
			}
		}

		if !found {
			// Try to find among connections
			for _, conn := range n.connections {
				if conn.tgt == ip.addr {
					channel = conn.channel
					found = true
					break
				}
			}
		}

		if !found {
			// Try to find a proc and attach a new channel to it
			recvPort, err := n.getProcPort(ip.addr.proc, ip.addr.port, reflect.RecvDir)
			if err != nil {
				return err
			}

			channel, err = attachPort(recvPort, ip.addr, reflect.RecvDir, reflect.ValueOf(nil), n.conf.BufferSize)
			found = true
			shouldClose = true
		}

		if found {
			// Send data to the port
			go func() {
				channel.Send(reflect.ValueOf(ip.data))
				if shouldClose {
					channel.Close()
				}
			}()
		} else {
			return fmt.Errorf("IIP target not found: '%s'", ip.addr)
		}
	}
	return nil
}
