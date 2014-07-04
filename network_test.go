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

// A constructor that can be used by component registry/factory
func newEchoer() interface{} {
	return new(echoer)
}

func init() {
	Register("echoer", newEchoer)
}

// A graph to test network features
type testNet struct {
	Graph

	InitTestFlag int
	FinTestFlag  chan bool
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
	// Exported state
	n.FinTestFlag = make(chan bool)
	return n
}

// Test for a network initializer
func (n *testNet) Init() {
	n.InitTestFlag = 123
}

// Test for a network finalizer
func (n *testNet) Finish() {
	n.InitTestFlag = 456
	n.FinTestFlag <- true
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

	// Run the test network
	RunNet(net)

	in <- 12
	i := <-out
	if i != 12 {
		t.Errorf("%d != %d", i, 12)
	}
	in <- net.InitTestFlag
	i = <-out
	if i != 123 {
		t.Errorf("After Init: %d != %d", i, 123)
	}

	close(in)
	// Wait for finalization signal
	<-net.FinTestFlag
	if net.InitTestFlag != 456 {
		t.Errorf("Finish: %d != %d", net.InitTestFlag, 456)
	}
	<-net.Wait()
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
	<-net.Wait()
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
	<-n.Wait()
}

// A struct to test IIPs support
type iipNet struct {
	Graph
}

// Creates a new test network with an IIP
func newIipNet() *iipNet {
	n := new(iipNet)
	n.InitGraphState()

	n.Add(new(echoer), "e1")

	n.AddIIP(interface{}(123), "e1", "In")

	n.MapInPort("In", "e1", "In")
	n.MapOutPort("Out", "e1", "Out")

	return n
}

// Tests IIP support in network
func TestIIP(t *testing.T) {
	net := newIipNet()
	in := make(chan int)
	out := make(chan int)
	net.SetInPort("In", in)
	net.SetOutPort("Out", out)

	RunNet(net)

	h := <-out
	if h != 123 {
		t.Errorf("%d != 123", h)
	}

	close(in)
	<-net.Wait()
}

// A simple syncrhonous summator for 2 arguments
type sum2 struct {
	Component

	Arg1 <-chan int
	Arg2 <-chan int
	Sum  chan<- int

	StateLock *sync.Mutex

	buf1 []int
	buf2 []int
}

func newSum2() *sum2 {
	s := new(sum2)
	s.StateLock = new(sync.Mutex)
	s.buf1 = make([]int, 0, 100)
	s.buf2 = make([]int, 0, 100)
	return s
}

func init() {
	Register("sum2", func() interface{} {
		return newSum2()
	})
}

// If available, pops arguments from the stack
// and sends the sum to the output
func (s *sum2) trySum() {
	if len(s.buf1) > 0 && len(s.buf2) > 0 {
		a1 := s.buf1[0]
		s.buf1 = s.buf1[1:]
		a2 := s.buf2[0]
		s.buf2 = s.buf2[1:]
		s.Sum <- (a1 + a2)
	}
}

func (s *sum2) OnArg1(a int) {
	s.buf1 = append(s.buf1, a)
	s.trySum()
}

func (s *sum2) OnArg2(a int) {
	s.buf2 = append(s.buf2, a)
	s.trySum()
}

// A network to test manual Stop() calls
type stopMeNet struct {
	Graph

	Fin chan int
}

func newStopMeNet() *stopMeNet {
	s := new(stopMeNet)
	s.InitGraphState()
	s.Fin = make(chan int)

	s.AddNew("doubler", "d1")
	s.AddNew("doubler", "d2")
	s.Connect("d1", "Out", "d2", "In")

	s.MapInPort("In", "d1", "In")
	s.MapOutPort("Out", "d2", "Out")
	return s
}

func (s *stopMeNet) Finish() {
	s.Fin <- 909
}

// Test manual network stopping method
func TestStopNet(t *testing.T) {
	s := newStopMeNet()
	in := make(chan int, 20)
	out := make(chan int, 20)
	s.SetInPort("In", in)
	s.SetOutPort("Out", out)

	RunNet(s)
	for i := 0; i < 10; i++ {
		in <- i
	}
	for i := 0; i < 10; i++ {
		i2 := <-out
		if i2 < 0 {
			t.Errorf("%d < 0", i2)
		}
	}
	// Stop without closing chans
	s.Stop()
	// Wait for finish signal
	fin := <-s.Fin
	if fin != 909 {
		t.Errorf("Invalid final signal: %d", fin)
	}
}

// type forked struct {
// 	Graph
// }

// func newForked() *forked {
// 	n := new(forked)
// 	n.InitGraphState()

// 	n.Add(new(echoer), "e1")
// 	n.Add(new(echoer), "e2")
// 	n.Add(newSum2(), "sum")

// 	n.Connect("e1", "Out", "sum", "Arg1")
// 	n.Connect("e2", "Out", "sum", "Arg2")

// 	n.MapInPort("In1", "e1", "In")
// 	n.MapInPort("In2", "e2", "In")
// 	n.MapOutPort("Out", "sum", "Sum")

// 	return n
// }

// func TestForkedNet(t *testing.T) {
// 	net := newForked()

// 	in1 := make(chan int)
// 	in2 := make(chan int)
// 	out := make(chan int)
// 	net.SetInPort("In1", in1)
// 	net.SetInPort("In2", in2)
// 	net.SetOutPort("Out", out)

// 	RunNet(net)

// 	in1 <- 2
// 	in2 <- 3

// 	i := <-out

// 	if i != 5 {
// 		t.Errorf("%d != 5\n", i)
// 	}

// 	close(in1)
// 	close(in2)

// 	<-net.Wait()
// }

// A graph to 1-to-N connections
type oneToNNet struct {
	Graph
}

func newOneToNNet(t *testing.T) *oneToNNet {
	// Initialization
	n := new(oneToNNet)
	n.InitGraphState()
	// Components
	e1 := new(echoer)
	e2 := new(echoer)
	e3 := new(echoer)
	// Structure
	if !n.Add(e1, "e1") {
		t.Errorf("Couldn't add e1")
	}
	if !n.Add(e2, "e2") {
		t.Errorf("Couldn't add e2")
	}
	if !n.Add(e3, "e3") {
		t.Errorf("Couldn't add e3")
	}
	if !n.Connect("e1", "Out", "e2", "In") {
		t.Errorf("net.Connect() returned false")
	}
	if !n.Connect("e1", "Out", "e3", "In") {
		t.Errorf("net.Connect() returned false")
	}
	// Ports
	n.MapInPort("In", "e1", "In")
	n.MapOutPort("Out2", "e2", "Out")
	n.MapOutPort("Out3", "e3", "Out")
	// Exported state
	return n
}

// Tests if 1-to-n connection work as they should in go
// i.e. we sond to multipe receivers and check if go pseudorandimly chooses receivers
func TestOneToNConnections(t *testing.T) {
	// Make the network of 2 components
	net := newOneToNNet(t)
	// in and out serve as network's in and out
	in := make(chan int)
	out2 := make(chan int)
	out3 := make(chan int)
	net.SetInPort("In", in)
	net.SetOutPort("Out2", out2)
	net.SetOutPort("Out3", out3)

	// Run the test network
	RunNet(net)

	var out2cnt, out3cnt uint
	for testnum := 12; testnum < 16; testnum++ {
		in <- testnum
		select {
		case i := <-out2:
			out2cnt++
			if i != testnum {
				t.Errorf("%d != %d", i, testnum)
			}
		case i := <-out3:
			out3cnt++
			if i != testnum {
				t.Errorf("%d != %d", i, testnum)
			}
		}
	}
	if out2cnt == 0 {
		t.Errorf("nothing was received on channel out2")
	}
	if out3cnt == 0 {
		t.Errorf("nothing was received on channel out3")
	}
	close(in)
	// Wait for finalization signal
	<-net.Wait()
}

// A graph to test N-to-1 connections
type nToOneNet struct {
	Graph
}

func newNToOneNet(t *testing.T) *nToOneNet {
	// Initialization
	n := new(nToOneNet)
	n.InitGraphState()
	// Components
	e1 := new(echoer)
	e2 := new(echoer)
	e3 := new(echoer)
	e4 := new(echoer)
	// Structure
	if !n.Add(e1, "e1") {
		t.Errorf("Couldn't add e1")
	}
	if !n.Add(e2, "e2") {
		t.Errorf("Couldn't add e2")
	}
	if !n.Add(e3, "e3") {
		t.Errorf("Couldn't add e3")
	}
	if !n.Add(e4, "e4") {
		t.Errorf("Couldn't add e3")
	}
	if !n.Connect("e1", "Out", "e4", "In") {
		t.Errorf("net.Connect() returned false")
	}
	if !n.Connect("e2", "Out", "e4", "In") {
		t.Errorf("net.Connect() returned false")
	}
	if !n.Connect("e3", "Out", "e4", "In") {
		t.Errorf("net.Connect() returned false")
	}
	// Ports
	n.MapInPort("In1", "e1", "In")
	n.MapInPort("In2", "e2", "In")
	n.MapInPort("In3", "e3", "In")
	n.MapOutPort("Out", "e4", "Out")
	// Exported state
	return n
}

// Tests if 1-to-n connection work as they should in go
// i.e. we sond to multipe receivers and check if go pseudorandimly chooses receivers
func TestNToOneConnections(t *testing.T) {
	// Make the network of 2 components
	net := newNToOneNet(t)
	// in and out serve as network's in and out
	in1 := make(chan int)
	in2 := make(chan int)
	in3 := make(chan int)
	out := make(chan int)
	net.SetInPort("In1", in1)
	net.SetInPort("In2", in2)
	net.SetInPort("In3", in3)
	net.SetOutPort("Out", out)

	// Run the test network
	RunNet(net)

	testnum := 12
	in1 <- testnum
	i := <-out
	if i != testnum {
		t.Errorf("%d != %d", i, testnum)
	}

	testnum = 24
	in2 <- testnum
	i = <-out
	if i != testnum {
		t.Errorf("%d != %d", i, testnum)
	}

	testnum = 36
	in3 <- testnum
	i = <-out
	if i != testnum {
		t.Errorf("%d != %d", i, testnum)
	}

	close(in1)
	close(in2)
	close(in3)
	//if we did not crash after closing, refounting worked

	// Wait for finalization signal
	<-net.Wait()
}
