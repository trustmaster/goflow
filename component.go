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
	var locker sync.Locker
	if lockField := v.FieldByName("StateLock"); lockField.IsValid() && lockField.Elem().IsValid() {
		locker, hasLock = lockField.Interface().(sync.Locker)
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
				// Add the input to the wait group
				inputsClose.Add(1)
				// Listen on an input channel
				go func() {
					for {
						val, ok := fv.Recv()
						if !ok {
							// The channel closed
							if hasClose {
								// Lock the state and call OnClose handler
								if hasLock {
									locker.Lock()
								}
								onClose.Call(empty)
								if hasLock {
									locker.Unlock()
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
									locker.Lock()
								}
								valArr := [1]reflect.Value{val}
								onRecv.Call(valArr[:])
								if hasLock {
									locker.Unlock()
								}
								handlersDone.Done()
							}()
						}
					}
				}()
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

		// Get the embedded flow.Component and check if it belongs to a network
		if vCom := v.FieldByName("Component"); vCom.IsValid() && vCom.Type().Name() == "Component" {
			if vNet := vCom.FieldByName("Net"); vNet.IsValid() && !vNet.IsNil() {
				if vNetCtr, hasNet := vNet.Interface().(netController); hasNet {
					// Remove the instance from the network's WaitGroup
					vNetCtr.getWait().Done()
				}
			}
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
