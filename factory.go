package flow

// DefaultRegistryCapacity is the capacity component registry is initialized with.
const DefaultRegistryCapacity = 64

// ComponentConstructor is a function that can be registered in the ComponentRegistry
// so that it is used when creating new processes of a specific component using
// Factory function at run-time.
type ComponentConstructor func(interface{}) interface{}

// ComponentRegistry is used to register components and spawn processes given just
// a string component name.
var ComponentRegistry = make(map[string]ComponentConstructor, DefaultRegistryCapacity)

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
func Factory(componentName string, initialPacket interface{}) interface{} {
	if constructor, exists := ComponentRegistry[componentName]; exists {
		return constructor(initialPacket)
	} else {
		panic("Uknown component name: " + componentName)
	}
}
