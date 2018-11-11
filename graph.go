package flow

import (
	"sync"
	"reflect"
)

type connection reflect.Value

// GraphConfig sets up properties for a graph
type GraphConfig struct {
	Capacity uint
	BufferSize uint
}

// defaultGraphConfig provides defaults for GraphConfig
func defaultGraphConfig() GraphConfig {
	return GraphConfig {
		Capacity: 32,
		BufferSize: 0,
	}
}

// Graph is a component that consists of other components connected with channels
type Graph struct {
	procs map[string]Component
	conns map[string]connection
	childGrp *sync.WaitGroup
}

// NewGraph returns a new initialized empty graph instance
func NewGraph(config ...GraphConfig) *Graph {
	conf := defaultGraphConfig()
	if (len(config) == 1) {
		conf = config[0]
	}

	return &Graph{
		procs: make(map[string]Component, conf.Capacity),
		conns: make(map[string]connection, conf.Capacity),
		childGrp: new(sync.WaitGroup),
	}
}

// Add a component to the graph
func (n *Graph) Add(name string, c Component) error {
	n.procs[name] = c
	return nil
}