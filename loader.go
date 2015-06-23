package flow

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"
	"strings"
)

// Internal representation of NoFlo JSON format
type GraphDescription struct {
	Properties struct {
		Name string
	}
	Processes map[string]struct {
		Component string
		Metadata  struct {
			Sync     bool  `json:",omitempty"`
			PoolSize int64 `json:",omitempty"`
		} `json:",omitempty"`
	}
	Connections []struct {
		Data interface{} `json:",omitempty"`
		Src  struct {
			Process string
			Port    string
		} `json:",omitempty"`
		Tgt struct {
			Process string
			Port    string
		}
		Metadata struct {
			Buffer int `json:",omitempty"`
		} `json:",omitempty"`
	}
	Exports []struct {
		Private string
		Public  string
	}
}

type parseJSONPair struct {
	Graph *Graph
	Error error
}

// ParseJSON converts a JSON network definition string into
// a flow.Graph object that can be run or used in other networks
func ParseJSON(js []byte) (*Graph, error) {
	r := parseJSON(js)
	return r.Graph, r.Error
}

func parseJSON(js []byte) (ret parseJSONPair) {
	// Parse JSON into Go struct
	var descr GraphDescription
	err := json.Unmarshal(js, &descr)
	if err != nil {
		ret.Error = err
		return
	}
	return create(&descr)
	// fmt.Printf("%+v\n", descr)

}

func Create(descr *GraphDescription) (*Graph, error) {
	r := create(descr)
	return r.Graph, r.Error
}

func create(descr *GraphDescription) (ret parseJSONPair) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			ret.Error = fmt.Errorf("%s", r)
		}
	}()

	constructor := func() interface{} {
		// Create a new Graph
		net := new(Graph)
		net.InitGraphState()

		// Add processes to the network
		for procName, procValue := range descr.Processes {
			net.AddNew(procValue.Component, procName)
			// Process mode detection
			if procValue.Metadata.PoolSize > 0 {
				proc := net.Get(procName).(*Component)
				proc.Mode = ComponentModePool
				proc.PoolSize = uint8(procValue.Metadata.PoolSize)
			} else if procValue.Metadata.Sync {
				proc := net.Get(procName).(*Component)
				proc.Mode = ComponentModeSync
			}
		}

		// Add connections
		for _, conn := range descr.Connections {
			// Check if it is an IIP or actual connection
			if conn.Data == nil {
				// Add a connection
				net.ConnectBuf(conn.Src.Process, conn.Src.Port, conn.Tgt.Process, conn.Tgt.Port, conn.Metadata.Buffer)
			} else {
				// Add an IIP
				net.AddIIP(conn.Data, conn.Tgt.Process, conn.Tgt.Port)
			}
		}

		// Add port exports
		for _, export := range descr.Exports {
			// Split private into proc.port
			procName := export.Private[:strings.Index(export.Private, ".")]
			procPort := export.Private[strings.Index(export.Private, ".")+1:]
			// Try to detect port direction using reflection
			procType := reflect.TypeOf(net.Get(procName)).Elem()
			field, fieldFound := procType.FieldByName(procPort)
			if !fieldFound {
				ret.Error = fmt.Errorf("Private port '%s' not found", export.Private)
				return nil
			}
			if field.Type.Kind() == reflect.Chan && (field.Type.ChanDir()&reflect.RecvDir) != 0 {
				// It's an inport
				net.MapInPort(export.Public, procName, procPort)
			} else if field.Type.Kind() == reflect.Chan && (field.Type.ChanDir()&reflect.SendDir) != 0 {
				// It's an outport
				net.MapOutPort(export.Public, procName, procPort)
			} else {
				// It's not a proper port
				ret.Error = fmt.Errorf("Private port '%s' not found", export.Private)
				return nil
			}
			// TODO add support for subgraphs
		}

		return net
	}

	// Register a component to be reused
	if descr.Properties.Name != "" {
		Register(descr.Properties.Name, constructor)
	}

	ret.Graph = constructor().(*Graph)
	return
}

// LoadJSON loads a JSON graph definition file into
// a flow.Graph object that can be run or used in other networks
func LoadJSON(filename string) *Graph {
	js, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}
	g, err := ParseJSON(js)
	if err != nil {
		// Hide errors per API
		return nil
	}
	return g
}

// RegisterJSON registers an external JSON graph definition as a component
// that can be instantiated at run-time using component Factory.
// It returns true on success or false if component name is already taken.
func RegisterJSON(componentName, filePath string) bool {
	var constructor ComponentConstructor
	constructor = func() interface{} {
		return LoadJSON(filePath)
	}
	return Register(componentName, constructor)
}
