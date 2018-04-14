// The flow package is a framework for Flow-based programming in Go.
package flow

import (
	"fmt"
	"reflect"
	"sync"
)

const (
	// ComponentModeUndefined stands for a fallback component mode (Async).
	ComponentModeUndefined = iota
	// ComponentModeAsync stands for asynchronous functioning mode.
	ComponentModeAsync
	// ComponentModeSync stands for synchronous functioning mode.
	ComponentModeSync
	// ComponentModePool stands for async functioning with a fixed pool.
	ComponentModePool
)

// DefaultComponentMode is the preselected functioning mode of all components being run.
var DefaultComponentMode = ComponentModeAsync

// Component is a generic flow component that has to be contained in concrete components.
// It stores network-specific information.
type Component struct {
	// Is running flag indicates that the process is currently running.
	IsRunning bool
	// Net is a pointer to network to inform it when the process is started and over
	// or to change its structure at run time.
	Net *Graph
	// Mode is component's functioning mode.
	Mode int8
	// PoolSize is used to define pool size when using ComponentModePool.
	PoolSize uint8
	// Term chan is used to terminate the process immediately without closing
	// any channels.
	Term chan struct{}
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

// Looper is a long-running process which actively receives data from its ports
// using a Loop function
type Looper interface {
	Loop()
}

// postHandler is used to bind handlers to a port
type portHandler struct {
	onRecv  reflect.Value
	onClose reflect.Value
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

	// Get the embedded flow.Component
	vCom := v.FieldByName("Component")
	isComponent := vCom.IsValid() && vCom.Type().Name() == "Component"

	if !isComponent {
		panic("Argument of flow.Run() is not a flow.Component")
	}

	// Get the component mode
	componentMode := DefaultComponentMode
	var poolSize uint8 = 0
	if vComMode := vCom.FieldByName("Mode"); vComMode.IsValid() {
		componentMode = int(vComMode.Int())
	}
	if vComPoolSize := vCom.FieldByName("PoolSize"); vComPoolSize.IsValid() {
		poolSize = uint8(vComPoolSize.Uint())
	}

	// Create a slice of select cases and port handlers
	cases := make([]reflect.SelectCase, 0, t.NumField())
	handlers := make([]portHandler, 0, t.NumField())

	// Make and listen on termination channel
	vCom.FieldByName("Term").Set(reflect.MakeChan(vCom.FieldByName("Term").Type(), 0))
	cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: vCom.FieldByName("Term")})
	handlers = append(handlers, portHandler{})

	// Detect active components
	looper, isLooper := c.(Looper)

	// Iterate over struct fields and bind handlers
	inputCount := 0
	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		ff := t.Field(i)
		ft := fv.Type()
		// Detect control channels
		if fv.IsValid() && fv.Kind() == reflect.Chan && !fv.IsNil() && fv.CanSet() && (ft.ChanDir()&reflect.RecvDir) != 0 {
			// Bind handlers for an input channel
			cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: fv})
			h := portHandler{onRecv: vp.MethodByName("On" + ff.Name), onClose: vp.MethodByName("On" + ff.Name + "Close")}
			handlers = append(handlers, h)
			if h.onClose.IsValid() || h.onRecv.IsValid() {
				// Add the input to the wait group
				inputsClose.Add(1)
				inputCount++
			}
		}
	}

	if inputCount == 0 && !isLooper {
		panic(fmt.Sprintf("Components with no input ports are not supported (%s)", reflect.TypeOf(c)))
	}

	// Prepare handler closures
	recvHandler := func(onRecv, value reflect.Value) {
		if hasLock {
			locker.Lock()
		}
		valArr := [1]reflect.Value{value}
		onRecv.Call(valArr[:])
		if hasLock {
			locker.Unlock()
		}
		handlersDone.Done()
	}
	closeHandler := func(onClose reflect.Value) {
		if onClose.IsValid() {
			// Lock the state and call OnClose handler
			if hasLock {
				locker.Lock()
			}
			onClose.Call([]reflect.Value{})
			if hasLock {
				locker.Unlock()
			}
		}
		inputsClose.Done()
	}
	terminate := func() {
		if !vCom.FieldByName("IsRunning").Bool() {
			return
		}
		vCom.FieldByName("IsRunning").SetBool(false)
		for i := 0; i < inputCount; i++ {
			inputsClose.Done()
		}
	}
	// closePorts closes all output channels of a process.
	closePorts := func() {
		// Iterate over struct fields
		for i := 0; i < t.NumField(); i++ {
			fv := v.Field(i)
			ft := fv.Type()
			vNet := vCom.FieldByName("Net")
			// Detect and close send-only channels
			if fv.IsValid() {
				// TODO: likely needs to check fv.CanSet()
				if fv.Kind() == reflect.Chan && (ft.ChanDir()&reflect.SendDir) != 0 && (ft.ChanDir()&reflect.RecvDir) == 0 {
					if vNet.IsValid() && !vNet.IsNil() {
						if vNet.Interface().(*Graph).DecSendChanRefCount(fv) {
							fv.Close()
						}
					} else {
						fv.Close()
					}
				} else if fv.Kind() == reflect.Slice && ft.Elem().Kind() == reflect.Chan {
					ll := fv.Len()
					if vNet.IsValid() && !vNet.IsNil() {
						for i := 0; i < ll; i += 1 {
							if vNet.Interface().(*Graph).DecSendChanRefCount(fv.Index(i)) {
								fv.Index(i).Close()
							}
						}
					} else {
						for i := 0; i < ll; i += 1 {
							fv.Index(i).Close()
						}
					}
				}
			}
		}
	}
	// shutdown represents a standard process shutdown procedure.
	shutdown := func() {
		if s, ok := c.(Shutdowner); ok {
			// Custom shutdown behavior
			s.Shutdown()
		} else {
			// Call user finish function if exists
			if finable, ok := c.(Finalizable); ok {
				finable.Finish()
			}
			// Close all output ports if the process is still running
			if vCom.FieldByName("IsRunning").Bool() {
				closePorts()
			}
		}
	}

	// This accomodates the looper behaviour specifically.
	// Because a looper does not rely on having a declared input handler, there is no blocking for inputsClosed.
	// This opens a race condition for handlersDone.
	handlersEst := make(chan bool, 1)

	// Run the port handlers depending on component mode
	if componentMode == ComponentModePool && poolSize > 0 {
		// Pool mode, prefork limited goroutine pool for all inputs
		var poolIndex uint8
		poolWait := new(sync.WaitGroup)
		once := new(sync.Once)
		for poolIndex = 0; poolIndex < poolSize; poolIndex++ {
			poolWait.Add(1)
			go func() {
				// TODO add pool of Loopers support
				for {
					chosen, recv, recvOK := reflect.Select(cases)
					if !recvOK {
						poolWait.Done()
						if chosen == 0 {
							// Term signal
							terminate()
						} else {
							// Port has been closed
							once.Do(func() {
								// Wait for other workers
								poolWait.Wait()
								// Close output down
								closeHandler(handlers[chosen].onClose)
							})
						}
						return
					}
					if handlers[chosen].onRecv.IsValid() {
						handlersDone.Add(1)
						recvHandler(handlers[chosen].onRecv, recv)
					}
				}
			}()
		}
		handlersEst <- true
	} else {
		go func() {
			if isLooper {
				defer func() {
					terminate()
					handlersDone.Done()
				}()
				handlersDone.Add(1)
				handlersEst <- true
				looper.Loop()
				return
			}
			handlersEst <- true
			for {
				chosen, recv, recvOK := reflect.Select(cases)
				if !recvOK {
					if chosen == 0 {
						// Term signal
						terminate()
					} else {
						// Port has been closed
						closeHandler(handlers[chosen].onClose)
					}
					return
				}
				if handlers[chosen].onRecv.IsValid() {
					handlersDone.Add(1)
					if componentMode == ComponentModeAsync || componentMode == ComponentModeUndefined && DefaultComponentMode == ComponentModeAsync {
						// Async mode
						go recvHandler(handlers[chosen].onRecv, recv)
					} else {
						// Sync mode
						recvHandler(handlers[chosen].onRecv, recv)
					}
				}
			}
		}()
	}

	// Indicate the process as running
	<-handlersEst
	vCom.FieldByName("IsRunning").SetBool(true)

	go func() {
		// Wait for all inputs to be closed
		inputsClose.Wait()
		// Wait all inport handlers to finish their job
		handlersDone.Wait()

		// Call shutdown handler (user or default)
		shutdown()

		// Get the embedded flow.Component and check if it belongs to a network
		if vNet := vCom.FieldByName("Net"); vNet.IsValid() && !vNet.IsNil() {
			if vNetCtr, hasNet := vNet.Interface().(netController); hasNet {
				// Remove the instance from the network's WaitGroup
				vNetCtr.getWait().Done()
			}
		}
	}()
	return true
}

// StopProc terminates the process if it is running.
// It doesn't close any in or out ports of the process, so it can be
// replaced without side effects.
func StopProc(c interface{}) bool {
	// Check if passed interface is a valid pointer to struct
	v := reflect.ValueOf(c)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		panic("Argument of TermProc() is not a valid pointer")
		return false
	}
	vp := v
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		panic("Argument of TermProc() is not a valid pointer to structure. Got type: " + vp.Type().Name())
		return false
	}
	// Get the embedded flow.Component
	vCom := v.FieldByName("Component")
	isComponent := vCom.IsValid() && vCom.Type().Name() == "Component"
	if !isComponent {
		panic("Argument of TermProc() is not a flow.Component")
	}
	// Send the termination signal
	vCom.FieldByName("Term").Close()
	return true
}
