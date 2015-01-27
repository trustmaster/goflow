package flow

import (
	"code.google.com/p/go.net/websocket"
	//"github.com/nu7hatch/gouuid"
	//"log"
	//"net"
	//"net/http"
    //"fmt"
)

type nodeHandler func(*websocket.Conn, interface{})

// Runtime is a NoFlo-compatible runtime implementing the FBP protocol
type Node struct {
	id       string
	handlers map[string]protocolHandler
	ready    chan struct{}
	done     chan struct{}
}
