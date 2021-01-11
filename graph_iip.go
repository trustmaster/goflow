package goflow

import (
	"fmt"
	"reflect"
)

// iip is the Initial Information Packet.
// IIPs are delivered to process input ports on the network start.
type iip struct {
	data interface{}
	addr address
}

// AddIIP adds an Initial Information packet to the network.
func (n *Graph) AddIIP(processName, portName string, data interface{}) error {
	addr := parseAddress(processName, portName)

	if _, exists := n.procs[processName]; exists {
		n.iips = append(n.iips, iip{data: data, addr: addr})
		return nil
	}

	return fmt.Errorf("AddIIP: could not find '%s'", addr)
}

// RemoveIIP detaches an IIP from specific process and port.
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

// sendIIPs sends Initial Information Packets upon network start.
func (n *Graph) sendIIPs() error {
	// Send initial IPs
	for i := range n.iips {
		ip := n.iips[i]

		// Get the receiver port channel
		channel, found := n.channelByInPortAddr(ip.addr)

		if !found {
			channel, found = n.channelByConnectionAddr(ip.addr)
		}

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
		}

		if !found {
			return fmt.Errorf("IIP target not found: '%s'", ip.addr)
		}

		// Increase reference count for the channel
		n.incChanListenersCount(channel)

		// Send data to the port
		go func(channel, data reflect.Value) {
			channel.Send(data)

			if n.decChanListenersCount(channel) {
				channel.Close()
			}
		}(channel, reflect.ValueOf(ip.data))
	}

	return nil
}

// channelByInPortAddr returns a channel by address from the network inports.
func (n *Graph) channelByInPortAddr(addr address) (channel reflect.Value, found bool) {
	for i := range n.inPorts {
		if n.inPorts[i].addr == addr {
			return n.inPorts[i].channel, true
		}
	}

	return reflect.Value{}, false
}

// channelByConnectionAddr returns a channel by address from connections.
func (n *Graph) channelByConnectionAddr(addr address) (channel reflect.Value, found bool) {
	for i := range n.connections {
		if n.connections[i].tgt == addr {
			return n.connections[i].channel, true
		}
	}

	return reflect.Value{}, false
}
