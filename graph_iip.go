package flow

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
