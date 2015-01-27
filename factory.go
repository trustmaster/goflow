package flow

// DefaultRegistryCapacity is the capacity component registry is initialized with.
const DefaultRegistryCapacity = 64

// ComponentConstructor is a function that can be registered in the ComponentRegistry
// so that it is used when creating new processes of a specific component using
// Factory function at run-time.
type ComponentConstructor func() interface{}

// ComponentRegistry is used to register components and spawn processes given just
// a string component name.
var ComponentRegistry = make(map[string]ComponentConstructor, DefaultRegistryCapacity)

// ComponentDescription is component metadata for IDE
type ComponentDescription struct {
	Description string
	Icon        string
	Metadata    map[string]interface{}
}

// ComponentDescriptions contains component metadata used by IDE
var ComponentDescriptions = make(map[string]ComponentDescription, DefaultRegistryCapacity)

// Describe sets a component description and metadata to be used in IDE
func Describe(componentName string, description ComponentDescription) {
	ComponentDescriptions[componentName] = description
}

// Register registers a component so that it can be instantiated at run-time using component Factory.
// It returns true on success or false if component name is already taken.
func Register(componentName string, constructor ComponentConstructor) bool {
	if _, exists := ComponentRegistry[componentName]; exists {
		// Component already registered
		return false
	}
	ComponentRegistry[componentName] = constructor
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
	if constructor, exists := ComponentRegistry[componentName]; exists {
		return constructor()
	} else {
		panic("Uknown component name: " + componentName)
	}
}
