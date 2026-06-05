package types

// Definition is the serializable intermediate representation of an FBP graph.
type Definition struct {
	Processes   map[string]ProcessDef `json:"processes"`
	Connections []ConnectionDef       `json:"connections,omitempty"`
	IIPs        []IIPDef              `json:"iips,omitempty"`
	Exports     []ExportDef           `json:"exports,omitempty"`
}

// ProcessDef describes a named process and its component.
type ProcessDef struct {
	Name      string `json:"name"`
	Component string `json:"component"`
}

// Endpoint identifies a process port, with an optional array index.
type Endpoint struct {
	Process string `json:"process"`
	Port    string `json:"port"`
	Index   *int   `json:"index,omitempty"`
}

// ConnectionDef describes a directed connection between two endpoints.
type ConnectionDef struct {
	Src Endpoint `json:"src"`
	Tgt Endpoint `json:"tgt"`
}

// IIPDef describes an initial information packet sent to a target endpoint.
type IIPDef struct {
	Data any      `json:"data"`
	Tgt  Endpoint `json:"tgt"`
}

// ExportKind distinguishes exported in-ports from exported out-ports.
type ExportKind string

const (
	// ExportInPort marks an external input port.
	ExportInPort ExportKind = "inport"
	// ExportOutPort marks an external output port.
	ExportOutPort ExportKind = "outport"
)

// ExportDef describes a graph-level port export declaration.
type ExportDef struct {
	Kind   ExportKind `json:"kind"`
	Public string     `json:"public"`
	Proc   string     `json:"proc"`
	Port   string     `json:"port"`
}

// DefinitionResult packages a collected Definition with any accumulated parse errors.
type DefinitionResult struct {
	Definition Definition
	Errors     []error
}

// FragmentKind identifies the type of graph element represented by a Fragment.
type FragmentKind string

const (
	// FragmentProcess is a process declaration fragment.
	FragmentProcess FragmentKind = "process"
	// FragmentConnection is a connection fragment.
	FragmentConnection FragmentKind = "connection"
	// FragmentIIP is an IIP fragment.
	FragmentIIP FragmentKind = "iip"
	// FragmentExport is an export declaration fragment.
	FragmentExport FragmentKind = "export"
	// FragmentError is a parse error fragment.
	FragmentError FragmentKind = "error"
)

// Fragment represents a single parsed element that will be collected
// into the final Definition. This allows parsers to emit multiple
// fragments per statement (e.g., inline component + connection).
type Fragment struct {
	Kind       FragmentKind
	Process    *ProcessDef
	Connection *ConnectionDef
	IIP        *IIPDef
	Export     *ExportDef
	Err        *ParseError
}
