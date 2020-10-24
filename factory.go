package goflow

import "fmt"

// Constructor is used to create a component instance at run-time
type Constructor func() (interface{}, error)

// Annotation provides reference information about a component
// to graph designers and operators
type Annotation struct {
	// Description tells what the component does
	Description string
	// Icon name in Font Awesome is used for visualization
	Icon string
}

// registryEntry contains runtime information about a component
type registryEntry struct {
	// Constructor is a function that creates a component instance.
	// It is required for the factory to add components at run-time.
	Constructor Constructor
	// Run-time component description
	Info ComponentInfo
}

// FactoryConfig sets up properties of a Factory
type FactoryConfig struct {
	RegistryCapacity uint
}

func defaultFactoryConfig() FactoryConfig {
	return FactoryConfig{
		RegistryCapacity: 64,
	}
}

// Factory registers components and creates their instances at run-time
type Factory struct {
	registry map[string]registryEntry
}

// NewFactory creates a new component Factory instance
func NewFactory(config ...FactoryConfig) *Factory {
	conf := defaultFactoryConfig()
	if len(config) == 1 {
		conf = config[0]
	}

	return &Factory{
		registry: make(map[string]registryEntry, conf.RegistryCapacity),
	}
}

// Register registers a component so that it can be instantiated at run-time
func (f *Factory) Register(componentName string, constructor Constructor) error {
	if _, exists := f.registry[componentName]; exists {
		return fmt.Errorf("registry error: component '%s' already registered", componentName)
	}

	f.registry[componentName] = registryEntry{
		Constructor: constructor,
		Info: ComponentInfo{
			Name: componentName,
		},
	}

	return nil
}

// Annotate adds human-readable documentation for a component to the runtime
func (f *Factory) Annotate(componentName string, annotation Annotation) error {
	if _, exists := f.registry[componentName]; !exists {
		return fmt.Errorf("registry annotation error: component '%s' is not registered", componentName)
	}

	entry := f.registry[componentName]
	entry.Info.Description = annotation.Description
	entry.Info.Icon = annotation.Icon
	f.registry[componentName] = entry

	return nil
}

// Unregister removes a component with a given name from the component registry and returns true
// or returns false if no such component is registered.
func (f *Factory) Unregister(componentName string) error {
	if _, exists := f.registry[componentName]; exists {
		delete(f.registry, componentName)
		return nil
	}

	return fmt.Errorf("registry error: component '%s' is not registered", componentName)
}

// Create creates a new instance of a component registered under a specific name.
func (f *Factory) Create(componentName string) (interface{}, error) {
	if info, exists := f.registry[componentName]; exists {
		return info.Constructor()
	}

	return nil, fmt.Errorf("factory error: component '%s' does not exist", componentName)
}

// // UpdateComponentInfo extracts run-time information about a
// // component and its ports. It is called when an FBP protocol client
// // requests component information.
// func (f *Factory) UpdateComponentInfo(componentName string) bool {
// 	component, exists := f.registry[componentName]
// 	if !exists {
// 		return false
// 	}
// 	// A component instance is required to reflect its type and ports
// 	instance := component.Constructor()

// 	component.Info.Name = componentName

// 	portMap, isGraph := instance.(portMapper)
// 	if isGraph {
// 		// Is a subgraph
// 		component.Info.Subgraph = true
// 		inPorts := portMap.listInPorts()
// 		component.Info.InPorts = make([]PortInfo, len(inPorts))
// 		for key, value := range inPorts {
// 			if value.info.Id == "" {
// 				value.info.Id = key
// 			}
// 			if value.info.Type == "" {
// 				value.info.Type = value.channel.Elem().Type().Name()
// 			}
// 			component.Info.InPorts = append(component.Info.InPorts, value.info)
// 		}
// 		outPorts := portMap.listOutPorts()
// 		component.Info.OutPorts = make([]PortInfo, len(outPorts))
// 		for key, value := range outPorts {
// 			if value.info.Id == "" {
// 				value.info.Id = key
// 			}
// 			if value.info.Type == "" {
// 				value.info.Type = value.channel.Elem().Type().Name()
// 			}
// 			component.Info.OutPorts = append(component.Info.OutPorts, value.info)
// 		}
// 	} else {
// 		// Is a component
// 		component.Info.Subgraph = false
// 		v := reflect.ValueOf(instance).Elem()
// 		t := v.Type()
// 		component.Info.InPorts = make([]PortInfo, t.NumField())
// 		component.Info.OutPorts = make([]PortInfo, t.NumField())
// 		for i := 0; i < t.NumField(); i++ {
// 			f := t.Field(i)
// 			if f.Type.Kind() == reflect.Chan {
// 				required := true
// 				if f.Tag.Get("required") == "false" {
// 					required = false
// 				}
// 				addressable := false
// 				if f.Tag.Get("addressable") == "true" {
// 					addressable = true
// 				}
// 				port := PortInfo{
// 					Id:          f.Name,
// 					Type:        f.Type.Name(),
// 					Description: f.Tag.Get("description"),
// 					Addressable: addressable,
// 					Required:    required,
// 				}
// 				if (f.Type.ChanDir() & reflect.RecvDir) != 0 {
// 					component.Info.InPorts = append(component.Info.InPorts, port)
// 				} else if (f.Type.ChanDir() & reflect.SendDir) != 0 {
// 					component.Info.OutPorts = append(component.Info.OutPorts, port)
// 				}
// 			}
// 		}
// 	}
// 	return true
// }
