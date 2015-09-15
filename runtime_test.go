package flow

import (
	"golang.org/x/net/websocket"
	"testing"
)

var (
	r       *Runtime
	started bool
)

func ensureRuntimeStarted() {
	if !started {
		r = new(Runtime)
		r.Init("goflow")
		go r.Listen("localhost:13014")
		started = true
		<-r.Ready()
	}
}

// Tests runtime information support
func TestRuntimeGetRuntime(t *testing.T) {
	ensureRuntimeStarted()
	// Create a WebSocket client
	ws, err := websocket.Dial("ws://localhost:13014/", "", "http://localhost/")
	if err != nil {
		t.Error(err.Error())
	}
	// Send a runtime request and check the response
	if err = websocket.JSON.Send(ws, &Message{"runtime", "getruntime", nil}); err != nil {
		t.Error(err.Error())
	}
	var msg runtimeMessage
	if err = websocket.JSON.Receive(ws, &msg); err != nil {
		t.Error(err.Error())
		return
	}
	if msg.Protocol != "runtime" || msg.Command != "runtime" {
		t.Errorf("Invalid protocol (%s) or command (%s)", msg.Protocol, msg.Command)
		return
	}
	res := msg.Payload
	if res.Type != "goflow" {
		t.Errorf("Invalid protocol type: %s\n", res.Type)
	}
	if res.Version != "0.4" {
		t.Errorf("Invalid protocol version: %s\n", res.Version)
	}
	if len(res.Capabilities) != 5 {
		t.Errorf("Invalid number of supported capabilities: %v\n", res.Capabilities)
	}
	if res.Id == "" {
		t.Error("Runtime Id is empty")
	}
}
