// The flow package is a framework for Flow-based programming in Go.
package flow

import (
	"reflect"
	"sync"
)

// Component is a generic flow component that has to be contained in concrete components.
// It stores network-specific information.
type Component struct {
	// Net is a pointer to network to inform it when the process is started and over
	// or to change its structure at run time.
	Net *Graph
}

// Initalizable is the interface implemented by components/graphs with custom initialization code.
type Initializable interface {
	Init()
}

// Finalizable is the interface implemented by components/graphs with extra finalization code.
type Finalizable interface {
	Finish()
}

// Shutdowner is the interface implemented by components overriding default Shutdown() behavior.
type Shutdowner interface {
	Shutdown()
}

// RunProc runs event handling loop on component ports.
// It returns true on success or panics with error message and returns false on error.
func RunProc(c interface{}) bool {
	// Check if passed interface is a valid pointer to struct
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		panic("Argument of flow.Run() is not a valid pointer")
		return false
	}
	vp := v
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		panic("Argument of flow.Run() is not a valid pointer to structure. Got type: " + vp.Type().Name())
		return false
	}
	t := v.Type()

	// Get internal state lock if available
	hasLock := false
	var lockField, lockFieldElem, lockFunc, unlockFunc reflect.Value
	lockField = v.FieldByName("StateLock")
	if lockField.IsValid() {
		lockFieldElem = lockField.Elem()
		hasLock = lockFieldElem.IsValid() && lockFieldElem.Type().Name() == "Mutex"
	}
	if hasLock {
		lockFunc = lockField.MethodByName("Lock")
		unlockFunc = lockField.MethodByName("Unlock")
	}

	// Get the embedded flow.Component
	vCom := v.FieldByName("Component")
	hasComponent := vCom.IsValid() && vCom.Type().Name() == "Component"
	var vNet reflect.Value
	hasNet := false // indicates whether it is attached to a network
	var vNetCtr netController
	if hasComponent {
		vNet = vCom.FieldByName("Net")
		if vNet.IsValid() && !vNet.IsNil() {
			if vNetCtr, hasNet = vNet.Interface().(netController); hasNet {
				// Add an instance to the network's WaitGroup
				vNetCtr.getWait().Add(1)
			}
		}
	}

	// Call user init function if exists
	if initable, ok := c.(Initializable); ok {
		initable.Init()
	}

	// A group to wait for all inputs to be closed
	inputsClose := new(sync.WaitGroup)
	// A group to wait for all recv handlers to finish
	handlersDone := new(sync.WaitGroup)

	emptyArr := [0]reflect.Value{}
	empty := emptyArr[:]
	
	// Bind channel event handlers
	// Iterate over struct fields
	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		ff := t.Field(i)
		ft := fv.Type()
		// Detect control channels
		if fv.IsValid() && fv.Kind() == reflect.Chan && !fv.IsNil() && (ft.ChanDir()&reflect.RecvDir) != 0 {
			// Bind handlers for an input channel
			onClose := vp.MethodByName("On" + ff.Name + "Close")
			hasClose := onClose.IsValid()
			onRecv := vp.MethodByName("On" + ff.Name)
			hasRecv := onRecv.IsValid()
			if hasClose || hasRecv {
				// Listen on an input channel
				go func() {
					for {
						val, ok := fv.Recv()
						if !ok {
							// The channel closed
							if hasClose {
								// Lock the state and call OnClose handler
								if hasLock {
									lockFunc.Call(empty)
								}
								onClose.Call(empty)
								if hasLock {
									unlockFunc.Call(empty)
								}
							}
							inputsClose.Done()
							return
						}
						if hasRecv {
							// Call the receival handler for this channel
							handlersDone.Add(1)
							go func() {
								if hasLock {
									lockFunc.Call(empty)
								}
								valArr := [1]reflect.Value{val}
								onRecv.Call(valArr[:])
								if hasLock {
									unlockFunc.Call(empty)
								}
								handlersDone.Done()
							}()
						}
					}
				}()
				// Add it to the wait group
				inputsClose.Add(1)
			}
		}
	}
	go func() {
		// Wait for all inputs to be closed
		inputsClose.Wait()
		// Wait all inport handlers to finish their job
		handlersDone.Wait()
		// Call shutdown handler (user or default)
		shutdownProc(c)
		// Remove the instance from the network's WaitGroup
		if hasNet {
			vNetCtr.getWait().Done()
		}
	}()
	return true
}

// closePorts closes all output channels of a process.
func closePorts(c interface{}) {
	v := reflect.ValueOf(c).Elem()
	t := v.Type()
	// Iterate over struct fields
	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		ft := fv.Type()
		// Detect and close send-only channels
		if fv.IsValid() && fv.Kind() == reflect.Chan && (ft.ChanDir()&reflect.SendDir) != 0 && (ft.ChanDir()&reflect.RecvDir) == 0 {
			fv.Close()
		}
	}
}

// shutdownProc represents a standard process shutdown procedure.
func shutdownProc(c interface{}) {
	if s, ok := c.(Shutdowner); ok {
		// Custom shutdown behavior
		s.Shutdown()
	} else {
		// Call user finish function if exists
		if finable, ok := c.(Finalizable); ok {
			finable.Finish()
		}
		// Close all output ports
		closePorts(c)
	}
}
