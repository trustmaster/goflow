package flow

import (
	"sync"
	"testing"
	"time"
)

// A component that doubles its int input
type doubler struct {
	Component
	In  <-chan int
	Out chan<- int
}

// Doubles the input and sends it to output
func (d *doubler) OnIn(i int) {
	d.Out <- i * 2
}

// A constructor that can be used by component registry/factory
func newDoubler() interface{} {
	return new(doubler)
}

func init() {
	Register("doubler", newDoubler)
}

// Tests a component with single input and single output
func TestSingleInput(t *testing.T) {
	d := new(doubler)
	in := make(chan int, 10)
	out := make(chan int, 10)
	d.In = in
	d.Out = out
	RunProc(d)
	for i := 0; i < 10; i++ {
		in <- i
		i2 := <-out
		ix2 := i * 2
		if i2 != ix2 {
			t.Errorf("%d != %d", i2, ix2)
		}
	}
	// Shutdown the component
	close(in)
}

// A component that locks to preserve concurrent modification of its state
type locker struct {
	Component
	In  <-chan int
	Out chan<- int

	StateLock *sync.Mutex

	counter int
	sum     int
}

// Creates a locker instance. This is required because StateLock must be a pointer
func newLocker() *locker {
	l := new(locker)
	l.counter = 0
	l.sum = 0
	l.StateLock = new(sync.Mutex)
	return l
}

// A constructor that can be used by component registry/factory
func newLockerConstructor() interface{} {
	return newLocker()
}

func init() {
	Register("locker", newLockerConstructor)
}

// Simulates long processing and read/write access
func (l *locker) OnIn(i int) {
	l.counter++
	// Half of the calls will wait to simulate long processing
	if l.counter%2 == 0 {
		time.Sleep(1000)
	}

	// Parellel write data race danger is here
	l.sum += i
}

func (l *locker) Shutdown() {
	// Emit the result and don't close the outport
	l.Out <- l.sum
}

// Tests internal state locking feature.
// Run with GOMAXPROCS > 1.
func TestStateLock(t *testing.T) {
	l := newLocker()
	in := make(chan int, 10)
	out := make(chan int, 10)
	l.In = in
	l.Out = out
	RunProc(l)
	// Simulate parallel writing and count the sum
	sum := 0
	for i := 1; i <= 1000; i++ {
		in <- i
		sum += i
	}
	// Send the close signal
	close(in)
	// Get the result and check if it is consistent
	sum2 := <-out
	if sum2 != sum {
		t.Errorf("%d != %d", sum2, sum)
	}
}

// Similar to locker, but intended to test ComponentModeSync
type syncLocker struct {
	Component
	In  <-chan int
	Out chan<- int

	counter int
	sum     int
}

// Creates a syncLocker instance
func newSyncLocker() *syncLocker {
	l := new(syncLocker)
	l.counter = 0
	l.sum = 0
	l.Component.Mode = ComponentModeSync // Change this to ComponentModeAsync and the test will fail
	return l
}

// A constructor that can be used by component registry/factory
func newSyncLockerConstructor() interface{} {
	return newSyncLocker()
}

func init() {
	Register("syncLocker", newSyncLockerConstructor)
}

// Simulates long processing and read/write access
func (l *syncLocker) OnIn(i int) {
	l.counter++
	// Half of the calls will wait to simulate long processing
	if l.counter%2 == 0 {
		time.Sleep(1000)
	}

	// Parellel write data race danger is here
	l.sum += i
}

func (l *syncLocker) Shutdown() {
	// Emit the result and don't close the outport
	l.Out <- l.sum
}

// Tests synchronous process execution feature.
// Run with GOMAXPROCS > 1.
func TestSyncLock(t *testing.T) {
	l := newSyncLocker()
	in := make(chan int, 10)
	out := make(chan int, 10)
	l.In = in
	l.Out = out
	RunProc(l)
	// Simulate parallel writing and count the sum
	sum := 0
	for i := 1; i <= 1000; i++ {
		in <- i
		sum += i
	}
	// Send the close signal
	close(in)
	// Get the result and check if it is consistent
	sum2 := <-out
	if sum2 != sum {
		t.Errorf("%d != %d", sum2, sum)
	}
}

// An external variable
var testInitFinFlag int

// Simple component
type initfin struct {
	Component
	In  <-chan int
	Out chan<- int
}

// Echo input
func (i *initfin) OnIn(n int) {
	// Dependent behavior
	if testInitFinFlag == 123 {
		i.Out <- n * 2
	} else {
		i.Out <- n
	}
}

// Initialization code, affects a global var
func (i *initfin) Init() {
	testInitFinFlag = 123
}

// Finalization code
func (i *initfin) Finish() {
	testInitFinFlag = 456
}

// Tests user initialization and finalization functions
func TestInitFinish(t *testing.T) {
	// Create and run the component
	i := new(initfin)
	i.Net = new(Graph)
	i.Net.InitGraphState()
	i.Net.waitGrp.Add(1)
	in := make(chan int)
	out := make(chan int)
	i.In = in
	i.Out = out
	RunProc(i)
	// Pass a value, the result must be affected by flag state
	in <- 2
	n2 := <-out
	if n2 != 4 {
		t.Errorf("%d != %d", n2, 4)
	}
	// Shut the component down and wait for Finish() code
	close(in)
	i.Net.waitGrp.Wait()
	if testInitFinFlag != 456 {
		t.Errorf("%d != %d", testInitFinFlag, 456)
	}
}

// A flag to test OnClose
var closeTestFlag int

// A component to test OnClose handlers
type closeTest struct {
	Component
	In <-chan int
}

// In channel close event handler
func (c *closeTest) OnInClose() {
	closeTestFlag = 789
}

// Tests close handler of input ports
func TestClose(t *testing.T) {
	c := new(closeTest)
	c.Net = new(Graph)
	c.Net.InitGraphState()
	c.Net.waitGrp.Add(1)
	in := make(chan int)
	c.In = in
	RunProc(c)
	in <- 1
	close(in)
	c.Net.waitGrp.Wait()
	if closeTestFlag != 789 {
		t.Errorf("%d != %d", closeTestFlag, 789)
	}
}

// A flag to test OnClose
var shutdownTestFlag int

// A component to test OnClose handlers
type shutdownTest struct {
	Component
	In <-chan int
}

// In channel close event handler
func (s *shutdownTest) OnIn(i int) {
	shutdownTestFlag = i
}

// Custom shutdown handler
func (s *shutdownTest) Shutdown() {
	shutdownTestFlag = 789
}

// Tests close handler of input ports
func TestShutdown(t *testing.T) {
	s := new(shutdownTest)
	s.Net = new(Graph)
	s.Net.InitGraphState()
	s.Net.waitGrp.Add(1)
	in := make(chan int)
	s.In = in
	RunProc(s)
	in <- 1
	close(in)
	s.Net.waitGrp.Wait()
	if shutdownTestFlag != 789 {
		t.Errorf("%d != %d", shutdownTestFlag, 789)
	}
}

func TestPoolMode(t *testing.T) {
	d := new(doubler)
	d.Component.Mode = ComponentModePool
	d.Component.PoolSize = 4
	in := make(chan int, 20)
	out := make(chan int, 20)
	d.In = in
	d.Out = out
	RunProc(d)
	for i := 0; i < 10; i++ {
		in <- i
	}
	for i := 0; i < 10; i++ {
		i2 := <-out
		if i2 < 0 {
			t.Errorf("%d < 0", i2)
		}
	}
	// Shutdown the component
	close(in)
}

// A component to test manual termination
type stopMe struct {
	Component
	In  <-chan int
	Out chan<- int
}

func (s *stopMe) OnIn(i int) {
	s.Out <- i * 2
}

func (s *stopMe) Finish() {
	s.Out <- 909
}

// Tests manual termination via StopProc()
func TestStopProc(t *testing.T) {
	s := new(stopMe)
	in := make(chan int, 20)
	out := make(chan int, 20)
	s.In = in
	s.Out = out
	// Test normal mode first
	RunProc(s)
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
	StopProc(s)
	// Wait for finish signal
	fin := <-out
	if fin != 909 {
		t.Errorf("Invalid final signal: %d", fin)
	}
	// Run again in Pool mode
	s.Component.Mode = ComponentModePool
	s.Component.PoolSize = 4
	RunProc(s)
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
	StopProc(s)
	// Wait for finish signal
	fin = <-out
	if fin != 909 {
		t.Errorf("Invalid final signal: %d", fin)
	}
}
