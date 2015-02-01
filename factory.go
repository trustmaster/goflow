package flow

import (
	"reflect"
)

// DefaultRegistryCapacity is the capacity component registry is initialized with.
const DefaultRegistryCapacity = 64

// ComponentConstructor is a function that can be registered in the ComponentRegistry
// so that it is used when creating new processes of a specific component using
// Factory function at run-time.
type ComponentConstructor func() interface{}

// ComponentEntry contains runtime information about a component
type ComponentEntry struct {
	// Constructor is a function that creates a component instance.
	// It is required for the factory to add components at run-time.
	Constructor ComponentConstructor
	// Run-time component description
	Info ComponentInfo
}

// ComponentRegistry is used to register components and spawn processes given just
// a string component name.
var ComponentRegistry = make(map[string]ComponentEntry, DefaultRegistryCapacity)

// Register registers a component so that it can be instantiated at run-time using component Factory.
// It returns true on success or false if component name is already taken.
func Register(componentName string, constructor ComponentConstructor) bool {
	if _, exists := ComponentRegistry[componentName]; exists {
		// Component already registered
		return false
	}
	ComponentRegistry[componentName] = ComponentEntry{
		Constructor: constructor,
	}
	return true
}

// Annotate sets component information utilized by runtimes and FBP protocol
// clients. Recommended fields are: Description and Icon. Other fields
// are infered by the runtime itself.
func Annotate(componentName string, info ComponentInfo) bool {
	component, exists := ComponentRegistry[componentName]
	if !exists {
		return false
	}
	component.Info = info
	return true
}

// Unregister removes a component with a given name from the component registry and returns true
// or returns false if no such component is registered.
func Unregister(componentName string) bool {
	if _, exists := ComponentRegistry[componentName]; exists {
		delete(ComponentRegistry, componentName)
		return true
	} else {
		return false
	}
}

// Factory creates a new instance of a component registered under a specific name.
func Factory(componentName string) interface{} {
	if info, exists := ComponentRegistry[componentName]; exists {
		return info.Constructor()
	} else {
		panic("Uknown component name: " + componentName)
	}
}

// UpdateComponentInfo extracts run-time information about a
// component and its ports. It is called when an FBP protocol client
// requests component information.
func UpdateComponentInfo(componentName string) bool {
	component, exists := ComponentRegistry[componentName]
	if !exists {
		return false
	}
	// A component instance is required to reflect its type and ports
	instance := component.Constructor()

	component.Info.Name = componentName

	portMap, isGraph := instance.(portMapper)
	if isGraph {
		// Is a subgraph
		component.Info.Subgraph = true
		inPorts := portMap.listInPorts()
		component.Info.InPorts = make([]PortInfo, len(inPorts))
		for key, value := range inPorts {
			if value.info.Id == "" {
				value.info.Id = key
			}
			if value.info.Type == "" {
				value.info.Type = value.channel.Elem().Type().Name()
			}
			component.Info.InPorts = append(component.Info.InPorts, value.info)
		}
		outPorts := portMap.listOutPorts()
		component.Info.OutPorts = make([]PortInfo, len(outPorts))
		for key, value := range outPorts {
			if value.info.Id == "" {
				value.info.Id = key
			}
			if value.info.Type == "" {
				value.info.Type = value.channel.Elem().Type().Name()
			}
			component.Info.OutPorts = append(component.Info.OutPorts, value.info)
		}
	} else {
		// Is a component
		component.Info.Subgraph = false
		v := reflect.ValueOf(instance).Elem()
		t := v.Type()
		component.Info.InPorts = make([]PortInfo, t.NumField())
		component.Info.OutPorts = make([]PortInfo, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Type.Kind() == reflect.Chan {
				required := true
				if f.Tag.Get("required") == "false" {
					required = false
				}
				addressable := false
				if f.Tag.Get("addressable") == "true" {
					addressable = true
				}
				port := PortInfo{
					Id:          f.Name,
					Type:        f.Type.Name(),
					Description: f.Tag.Get("description"),
					Addressable: addressable,
					Required:    required,
				}
				if (f.Type.ChanDir() & reflect.RecvDir) != 0 {
					component.Info.InPorts = append(component.Info.InPorts, port)
				} else if (f.Type.ChanDir() & reflect.SendDir) != 0 {
					component.Info.OutPorts = append(component.Info.OutPorts, port)
				}
			}
		}
	}
	return true
}
