package flow

import (
	"sync"
	"testing"
)

// Trivial component
type echoer struct {
	Component
	In  <-chan int
	Out chan<- int
}

// Sends recvd value to out
func (e *echoer) OnIn(i int) {
	e.Out <- i
}

var initTestFlag int
var finTestFlag chan bool

// A graph to test network features
type testNet struct {
	Graph
}

func newTestNet(t *testing.T) *testNet {
	// Initialization
	n := new(testNet)
	n.InitGraphState()
	// Components
	e1 := new(echoer)
	e2 := new(echoer)
	// Structure
	if !n.Add(e1, "e1") {
		t.Errorf("Couldn't add e1")
	}
	if !n.Add(e2, "e2") {
		t.Errorf("Couldn't add e2")
	}
	if !n.Connect("e1", "Out", "e2", "In") {
		t.Errorf("net.Connect() returned false")
	}
	// Ports
	n.MapInPort("In", "e1", "In")
	n.MapOutPort("Out", "e2", "Out")
	return n
}

// Test for a network initializer
func (n *testNet) Init() {
	initTestFlag = 123
}

// Test for a network finalizer
func (n *testNet) Finish() {
	initTestFlag = 456
	finTestFlag <- true
}

// Tests a simple connection between two components in the net
// and network initialization/finalization handlers
func TestConnection(t *testing.T) {
	// Make the network of 2 components
	net := newTestNet(t)
	// in and out serve as network's in and out
	in := make(chan int)
	out := make(chan int)
	net.SetInPort("In", in)
	net.SetOutPort("Out", out)

	// Finalization is captured by this channel
	finTestFlag = make(chan bool)

	// Run the test network
	RunNet(net)

	in <- 12
	i := <-out
	if i != 12 {
		t.Errorf("%d != %d", i, 12)
	}
	in <- initTestFlag
	i = <-out
	if i != 123 {
		t.Errorf("After Init: %d != %d", i, 123)
	}

	close(in)
	// Wait for finalization signal
	<-finTestFlag
	if initTestFlag != 456 {
		t.Errorf("Finish: %d != %d", initTestFlag, 456)
	}
}

// Structure to test 2-level composition
type compositeTest struct {
	Graph
}

// Creates a composite with processes and subnets
func newCompositeTest(t *testing.T) *compositeTest {
	// Initialization
	n := new(compositeTest)
	n.InitGraphState()
	// Structure
	if !n.Add(new(echoer), "e1") {
		t.Errorf("Couldn't add e1")
	}
	if !n.Add(newTestNet(t), "sub1") {
		t.Errorf("Couldn't add sub")
	}
	if !n.Add(newTestNet(t), "sub2") {
		t.Errorf("Couldn't add sub")
	}
	if !n.Add(newTestNet(t), "sub3") {
		t.Errorf("Couldn't add sub")
	}
	if !n.Connect("sub1", "Out", "e1", "In") {
		t.Errorf("net.Connect() returned false")
	}
	if !n.Connect("e1", "Out", "sub2", "In") {
		t.Errorf("net.Connect() returned false")
	}
	if !n.Connect("sub2", "Out", "sub3", "In") {
		t.Errorf("net.Connect() returned false")
	}
	// Ports
	n.MapInPort("In", "sub1", "In")
	n.MapOutPort("Out", "sub3", "Out")
	return n
}

// Tests a composite with processes and subnets
func TestComposite(t *testing.T) {
	// Make the network
	net := newCompositeTest(t)
	// in and out serve as network's in and out
	in := make(chan int)
	out := make(chan int)
	net.SetInPort("In", in)
	net.SetOutPort("Out", out)

	// Run the test network
	RunNet(net)

	in <- 42
	i := <-out
	if i != 42 {
		t.Errorf("%d != %d", i, 42)
	}

	close(in)
}

type rr struct {
	In  <-chan int
	Out []chan<- int

	StateLock *sync.Mutex

	Component
	idx int
}

func (r *rr) OnIn(i int) {
	pick := r.idx
	r.idx = (r.idx + 1) % len(r.Out)

	r.Out[pick] <- i
}

/*
 * Creates a simple network with a load balancer that round robins to its out
 * channels. Then sends to messages in and expects a response, 1 from each
 * of the out channels.
 */
func TestMultiOutChannel(t *testing.T) {
	n := new(compositeTest)
	n.InitGraphState()

	r := new(rr)
	if !n.Add(r, "lb") {
		t.Error("Unable to add load balancer")
	}

	e1 := new(echoer)
	if !n.Add(e1, "e1") {
		t.Error("Unable to add second echoer, e1")
	}

	e2 := new(echoer)
	if !n.Add(e2, "e2") {
		t.Error("Unable to add second echoer, e2")
	}

	if !n.Connect("lb", "Out", "e1", "In") {
		t.Error("Unable to connect LB to e1")
	}

	if !n.Connect("lb", "Out", "e2", "In") {
		t.Error("Unable to connect LB to e2")
	}

	if !n.MapInPort("In", "lb", "In") {
		t.Error("Unable to map InPort")
	}
	if !n.MapOutPort("Out1", "e1", "Out") {
		t.Error("Unable to mape OutPort 1")
	}

	if !n.MapOutPort("Out2", "e2", "Out") {
		t.Error("Unable to mape OutPort 2")
	}

	in := make(chan int)
	out1 := make(chan int)
	out2 := make(chan int)
	n.SetInPort("In", in)
	n.SetOutPort("Out1", out1)
	n.SetOutPort("Out2", out2)
	RunNet(n)

	in <- 42
	i := <-out1
	if i != 42 {
		t.Errorf("%d != %d", i, 42)
	}

	in <- 42
	i = <-out2
	if i != 42 {
		t.Errorf("%d != %d", i, 42)
	}

	close(in)
}
