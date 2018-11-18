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

// // run runs the network and waits for all processes to finish.
// func (n *Graph) run() {
// 	// Send initial IPs
// 	for _, ip := range n.iips {
// 		// Get the reciever port
// 		var rport reflect.Value
// 		found := false

// 		// Try to find it among network inports
// 		for _, inPort := range n.inPorts {
// 			if inPort.proc == ip.proc && inPort.port == ip.port {
// 				rport = inPort.channel
// 				found = true
// 				break
// 			}
// 		}

// 		if !found {
// 			// Try to find among connections
// 			for _, conn := range n.connections {
// 				if conn.tgt.proc == ip.proc && conn.tgt.port == ip.port {
// 					rport = conn.channel
// 					found = true
// 					break
// 				}
// 			}
// 		}

// 		if !found {
// 			// Try to find a proc and attach a new channel to it
// 			for procName, proc := range n.procs {
// 				if procName == ip.proc {
// 					// Check if receiver is a net
// 					rv := reflect.ValueOf(proc).Elem()
// 					var rnet reflect.Value
// 					if rv.Type().Name() == "Graph" {
// 						rnet = rv
// 					} else {
// 						rnet = rv.FieldByName("Graph")
// 					}
// 					if rnet.IsValid() {
// 						if pm, isPm := rnet.Addr().Interface().(portMapper); isPm {
// 							rport = pm.getInPort(ip.port)
// 						}
// 					} else {
// 						// Receiver is a proc
// 						rport = rv.FieldByName(ip.port)
// 					}

// 					// Validate receiver port
// 					rtport := rport.Type()
// 					if rtport.Kind() != reflect.Chan || rtport.ChanDir()&reflect.RecvDir == 0 {
// 						panic(ip.proc + "." + ip.port + " is not a valid input channel")
// 					}
// 					var channel reflect.Value

// 					// Make a channel of an appropriate type
// 					chanType := reflect.ChanOf(reflect.BothDir, rtport.Elem())
// 					channel = reflect.MakeChan(chanType, DefaultBufferSize)
// 					// Set the channel
// 					if rport.CanSet() {
// 						rport.Set(channel)
// 					} else {
// 						panic(ip.proc + "." + ip.port + " is not settable")
// 					}

// 					// Use the new channel to send the IIP
// 					rport = channel
// 					found = true
// 					break
// 				}
// 			}
// 		}

// 		if found {
// 			// Send data to the port
// 			rport.Send(reflect.ValueOf(ip.data))
// 		} else {
// 			panic("IIP target not found: " + ip.proc + "." + ip.port)
// 		}
// 	}
// }
