package goflow

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
)

// Internal representation of NoFlo JSON format
type graphDescription struct {
	Properties struct {
		Name string
	}
	Processes map[string]struct {
		Component string
		Metadata  struct {
			Sync     bool  `json:",omitempty"`
			PoolSize int64 `json:",omitempty"`
		} `json:",omitempty"`
		Parameters map[string]interface{} `json:",omitempty"`
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

// ParseJSON converts a JSON network definition string into
// a flow.Graph object that can be run or used in other networks
func ParseJSON(js []byte, factory *Factory) *Graph {
	// Parse JSON into Go struct
	var descr graphDescription
	err := json.Unmarshal(js, &descr)
	if err != nil {
		fmt.Println("Error parsing JSON", err)
		return nil
	}
	// fmt.Printf("%+v\n", descr)

	// Create a new Graph
	//net := new(Graph)
	//net.InitGraphState()
	net := NewGraph()

	// Add processes to the network
	for procName, procValue := range descr.Processes {
		net.AddNew(procName, procValue.Component, factory)
		proc, ok := net.Get(procName).(PlugIn)
		if ok {
			proc.SetParams(procValue.Parameters)
		}
		// Process mode detection
		// if procValue.Metadata.PoolSize > 0 {
		// 	proc := net.Get(procName).(*Component)
		// 	proc.Mode = ComponentModePool
		// 	proc.PoolSize = uint8(procValue.Metadata.PoolSize)
		// } else if procValue.Metadata.Sync {
		// 	proc := net.Get(procName).(*Component)
		// 	proc.Mode = ComponentModeSync
		// }
	}

	// Add connections
	for _, conn := range descr.Connections {
		// Check if it is an IIP or actual connection
		if conn.Data == nil {
			// Add a connection
			net.ConnectBuf(conn.Src.Process, conn.Src.Port, conn.Tgt.Process, conn.Tgt.Port, conn.Metadata.Buffer)
		} else {
			// Add an IIP
			net.AddIIP(conn.Tgt.Process, conn.Tgt.Port, conn.Data)
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
			panic("Private port '" + export.Private + "' not found")
		}
		if field.Type.Kind() == reflect.Chan && (field.Type.ChanDir()&reflect.RecvDir) != 0 {
			// It's an inport
			net.MapInPort(export.Public, procName, procPort)
		} else if field.Type.Kind() == reflect.Chan && (field.Type.ChanDir()&reflect.SendDir) != 0 {
			// It's an outport
			net.MapOutPort(export.Public, procName, procPort)
		} else {
			// It's not a proper port
			panic("Private port '" + export.Private + "' is not a valid channel")
		}
		// TODO add support for subgraphs
	}

	return net
}

// LoadJSON loads a JSON graph definition file into
// a flow.Graph object that can be run or used in other networks
func LoadJSON(filename string, factory *Factory) *Graph {
	js, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil
	}
	return ParseJSON(js, factory)
}

// RegisterJSON registers an external JSON graph definition as a component
// that can be instantiated at run-time using component Factory.
// It returns true on success or false if component name is already taken.
// func RegisterJSON(componentName, filePath string) bool {
// 	var constructor ComponentConstructor
// 	constructor = func() interface{} {
// 		return LoadJSON(filePath)
// 	}
// 	return Register(componentName, constructor)
// }
